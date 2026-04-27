package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"time"

	"probubbly/internal/auth"
	"probubbly/internal/db"
	"probubbly/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ResolveEvent(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	eventID := chi.URLParam(r, "id")

	var req models.ResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Outcome != "yes" && req.Outcome != "no" {
		http.Error(w, "Outcome must be 'yes' or 'no'", http.StatusBadRequest)
		return
	}

	var creatorID string
	var status string
	err := db.DB.QueryRow(db.Rebind(`
		SELECT creator_id, status FROM events WHERE id = ?`), eventID,
	).Scan(&creatorID, &status)

	if err == sql.ErrNoRows {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		return
	}

	isAdmin := auth.IsAdmin(r)
if !isAdmin {
    http.Error(w, "Only admin can resolve events", http.StatusForbidden)
    return
}

	if status != "open" {
		http.Error(w, "Event is already resolved", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction failed to start", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = tx.Exec(db.Rebind(`
		UPDATE events SET status = 'resolved', outcome = ?, resolved_at = ?
		WHERE id = ?`), req.Outcome, now, eventID,
	)
	if err != nil {
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
		return
	}

	rows, err := tx.Query(db.Rebind(`
		SELECT id, user_id, side, amount, ratio FROM predictions WHERE event_id = ?`), eventID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch predictions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type predictionData struct {
		id     string
		userID string
		side   string
		amount int
		ratio  float64
	}

	predictions := []predictionData{}
	for rows.Next() {
		var p predictionData
		if err := rows.Scan(&p.id, &p.userID, &p.side, &p.amount, &p.ratio); err != nil {
			continue
		}
		predictions = append(predictions, p)
	}
	rows.Close()

	totalHouseCut := 0.0
	winnersCount := 0
	losersCount := 0

	for _, p := range predictions {
		if p.side == req.Outcome {
			grossPayout := float64(p.amount) * p.ratio
			houseCut := grossPayout * 0.03
			netPayout := math.Round((grossPayout-houseCut)*10000) / 10000
			totalHouseCut += houseCut

			_, err = tx.Exec(db.Rebind(`
				UPDATE users SET balance = balance + ? WHERE id = ?`), netPayout, p.userID,
			)
			if err != nil {
				http.Error(w, "Failed to update winner balance", http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec(db.Rebind(`
				UPDATE predictions SET payout = ? WHERE id = ?`), netPayout, p.id,
			)
			if err != nil {
				http.Error(w, "Failed to update prediction payout", http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec(db.Rebind(`
				INSERT INTO transactions (id, user_id, type, amount, description, created_at)
				VALUES (?, ?, 'payout', ?, ?, ?)`),
				uuid.New().String(), p.userID, netPayout,
				"Won prediction on event ("+req.Outcome+")", now,
			)
			if err != nil {
				http.Error(w, "Failed to record payout transaction", http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec(db.Rebind(`
				INSERT INTO house_ledger (id, event_id, prediction_id, cut_amount, created_at)
				VALUES (?, ?, ?, ?, ?)`),
				uuid.New().String(), eventID, p.id, houseCut, now,
			)
			if err != nil {
				http.Error(w, "Failed to record house cut", http.StatusInternalServerError)
				return
			}

			winnersCount++
		} else {
			_, err = tx.Exec(db.Rebind(`
				UPDATE predictions SET payout = 0 WHERE id = ?`), p.id,
			)
			if err != nil {
				http.Error(w, "Failed to update prediction", http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec(db.Rebind(`
				INSERT INTO transactions (id, user_id, type, amount, description, created_at)
				VALUES (?, ?, 'loss', 0, ?, ?)`),
				uuid.New().String(), p.userID,
				"Lost prediction on event ("+req.Outcome+")", now,
			)
			if err != nil {
				http.Error(w, "Failed to record loss transaction", http.StatusInternalServerError)
				return
			}

			losersCount++
		}
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event_id":        eventID,
		"outcome":         req.Outcome,
		"winners":         winnersCount,
		"losers":          losersCount,
		"total_house_cut": totalHouseCut,
		"message":         "Event resolved successfully",
	})
}
