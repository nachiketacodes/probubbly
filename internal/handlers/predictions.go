package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"probubbly/internal/auth"
	"probubbly/internal/db"
	"probubbly/internal/models"
	"probubbly/internal/ratio"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func PlacePrediction(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	eventID := chi.URLParam(r, "id")

	var req models.PredictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Amount < 2 {
		http.Error(w, "Minimum prediction is 2 coins", http.StatusBadRequest)
		return
	}

	if req.Side != "yes" && req.Side != "no" {
		http.Error(w, "Side must be 'yes' or 'no'", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction failed to start", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var event models.Event
	var eventTime sql.NullString
	err = tx.QueryRow(db.Rebind(`
		SELECT id, title, status, yes_coins, no_coins, event_date, event_time
		FROM events WHERE id = ?`), eventID,
	).Scan(&event.ID, &event.Title, &event.Status, &event.YesCoins, &event.NoCoins, &event.EventDate, &eventTime)

	if err == sql.ErrNoRows {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		return
	}

	if eventTime.Valid {
		event.EventTime = eventTime.String
	}

	if event.Status != "open" {
		http.Error(w, "Event is not open for predictions", http.StatusBadRequest)
		return
	}

	var balance float64
	err = tx.QueryRow(db.Rebind("SELECT balance FROM users WHERE id = ?"), userID).Scan(&balance)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if balance < float64(req.Amount) {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	var totalUserBet int
	err = tx.QueryRow(db.Rebind(`
		SELECT COALESCE(SUM(amount), 0) FROM predictions
		WHERE event_id = ? AND user_id = ?`), eventID, userID,
	).Scan(&totalUserBet)
	if err != nil {
		http.Error(w, "Failed to check prediction limit", http.StatusInternalServerError)
		return
	}

	if totalUserBet+req.Amount > 80 {
		http.Error(w, "Maximum 80 coins per event exceeded", http.StatusBadRequest)
		return
	}

	currentRatios := ratio.Compute(event.YesCoins, event.NoCoins)
	lockedRatio := currentRatios.Yes
	if req.Side == "no" {
		lockedRatio = currentRatios.No
	}

	_, err = tx.Exec(db.Rebind("UPDATE users SET balance = balance - ? WHERE id = ?"), req.Amount, userID)
	if err != nil {
		http.Error(w, "Failed to deduct balance", http.StatusInternalServerError)
		return
	}

	if req.Side == "yes" {
		_, err = tx.Exec(db.Rebind("UPDATE events SET yes_coins = yes_coins + ? WHERE id = ?"), req.Amount, eventID)
	} else {
		_, err = tx.Exec(db.Rebind("UPDATE events SET no_coins = no_coins + ? WHERE id = ?"), req.Amount, eventID)
	}
	if err != nil {
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
		return
	}

	var username string
	err = tx.QueryRow(db.Rebind("SELECT username FROM users WHERE id = ?"), userID).Scan(&username)
	if err != nil {
		http.Error(w, "Failed to get username", http.StatusInternalServerError)
		return
	}

	predictionID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err = tx.Exec(db.Rebind(`
		INSERT INTO predictions (id, event_id, user_id, side, amount, ratio, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`),
		predictionID, eventID, userID, req.Side, req.Amount, lockedRatio, now,
	)
	if err != nil {
		http.Error(w, "Failed to create prediction", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(db.Rebind(`
		INSERT INTO transactions (id, user_id, type, amount, description, created_at)
		VALUES (?, ?, 'predict', ?, ?, ?)`),
		uuid.New().String(), userID, -req.Amount,
		"Predicted "+req.Side+" on \""+event.Title+"\"", now,
	)
	if err != nil {
		http.Error(w, "Failed to record transaction", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	newYesCoins := event.YesCoins
	newNoCoins := event.NoCoins
	if req.Side == "yes" {
		newYesCoins += req.Amount
	} else {
		newNoCoins += req.Amount
	}
	newRatios := ratio.Compute(newYesCoins, newNoCoins)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prediction": models.Prediction{
			ID:        predictionID,
			EventID:   eventID,
			UserID:    userID,
			Username:  username,
			Side:      req.Side,
			Amount:    req.Amount,
			Ratio:     lockedRatio,
			CreatedAt: now,
		},
		"new_balance": balance - float64(req.Amount),
		"new_ratios":  newRatios,
	})
}
