package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"probubbly/internal/db"
	"probubbly/internal/models"
)

func GetPlatformStats(w http.ResponseWriter, r *http.Request) {
	var totalUsers, totalEvents, openEvents, resolvedEvents, totalPredictions int
	var totalCoinsInPlay, totalHouseEarnings float64

	db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	db.DB.QueryRow("SELECT COUNT(*) FROM events").Scan(&totalEvents)
	db.DB.QueryRow("SELECT COUNT(*) FROM events WHERE status = 'open'").Scan(&openEvents)
	db.DB.QueryRow("SELECT COUNT(*) FROM events WHERE status = 'resolved'").Scan(&resolvedEvents)
	db.DB.QueryRow("SELECT COUNT(*) FROM predictions").Scan(&totalPredictions)
	db.DB.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM users").Scan(&totalCoinsInPlay)
	db.DB.QueryRow("SELECT COALESCE(SUM(cut_amount), 0) FROM house_ledger").Scan(&totalHouseEarnings)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_users":          totalUsers,
		"total_events":         totalEvents,
		"open_events":          openEvents,
		"resolved_events":      resolvedEvents,
		"total_predictions":    totalPredictions,
		"total_coins_in_play":  totalCoinsInPlay,
		"total_house_earnings": totalHouseEarnings,
	})
}

func ListAllUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, login_id, username, balance, borrowed, is_admin, joined_at
		FROM users ORDER BY joined_at DESC`)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []map[string]interface{}{}
	for rows.Next() {
		var id, loginID, username, joinedAt string
		var balance, borrowed float64
		var isAdmin int

		err := rows.Scan(&id, &loginID, &username, &balance, &borrowed, &isAdmin, &joinedAt)
		if err != nil {
			continue
		}

		var predictionCount int
		db.DB.QueryRow(db.Rebind("SELECT COUNT(*) FROM predictions WHERE user_id = ?"), id).Scan(&predictionCount)

		users = append(users, map[string]interface{}{
			"id":               id,
			"login_id":         loginID,
			"username":         username,
			"balance":          balance,
			"borrowed":         borrowed,
			"is_admin":         isAdmin == 1,
			"joined_at":        joinedAt,
			"prediction_count": predictionCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetUserDetail(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	var user models.User
	var lastBorrow sql.NullString
	var isAdmin int

	err := db.DB.QueryRow(db.Rebind(`
		SELECT id, login_id, username, balance, borrowed, last_borrow, is_admin, joined_at
		FROM users WHERE id = ?`), userID,
	).Scan(&user.ID, &user.LoginID, &user.Username, &user.Balance, &user.Borrowed, &lastBorrow, &isAdmin, &user.JoinedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		return
	}

	if lastBorrow.Valid {
		user.LastBorrow = lastBorrow.String
	}
	user.IsAdmin = isAdmin == 1

	rows, err := db.DB.Query(db.Rebind(`
		SELECT p.id, p.event_id, e.title, p.side, p.amount, p.ratio, p.payout, p.created_at
		FROM predictions p
		JOIN events e ON p.event_id = e.id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT 20`), userID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch predictions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	predictions := []map[string]interface{}{}
	for rows.Next() {
		var id, eventID, eventTitle, side, createdAt string
		var amount int
		var ratio, payout float64

		err := rows.Scan(&id, &eventID, &eventTitle, &side, &amount, &ratio, &payout, &createdAt)
		if err != nil {
			continue
		}

		predictions = append(predictions, map[string]interface{}{
			"id":          id,
			"event_id":    eventID,
			"event_title": eventTitle,
			"side":        side,
			"amount":      amount,
			"ratio":       ratio,
			"payout":      payout,
			"created_at":  createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":        user,
		"predictions": predictions,
	})
}

func GetHouseLedger(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT h.id, h.event_id, e.title, h.prediction_id, h.cut_amount, h.created_at
		FROM house_ledger h
		JOIN events e ON h.event_id = e.id
		ORDER BY h.created_at DESC
		LIMIT 100`)
	if err != nil {
		http.Error(w, "Failed to fetch house ledger", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	ledger := []map[string]interface{}{}
	for rows.Next() {
		var id, eventID, eventTitle, predictionID, createdAt string
		var cutAmount float64

		err := rows.Scan(&id, &eventID, &eventTitle, &predictionID, &cutAmount, &createdAt)
		if err != nil {
			continue
		}

		ledger = append(ledger, map[string]interface{}{
			"id":            id,
			"event_id":      eventID,
			"event_title":   eventTitle,
			"prediction_id": predictionID,
			"cut_amount":    cutAmount,
			"created_at":    createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ledger)
}
