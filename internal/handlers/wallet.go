package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"probubbly/internal/auth"
	"probubbly/internal/db"
	"probubbly/internal/models"

	"github.com/google/uuid"
)

func GetWallet(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	var lastBorrow sql.NullString
	var isAdmin int

	err := db.DB.QueryRow(db.Rebind(`
		SELECT id, username, balance, borrowed, last_borrow, is_admin, joined_at
		FROM users WHERE id = ?`), userID,
	).Scan(&user.ID, &user.Username, &user.Balance, &user.Borrowed, &lastBorrow, &isAdmin, &user.JoinedAt)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if lastBorrow.Valid {
		user.LastBorrow = lastBorrow.String
	} else {
		user.LastBorrow = ""
	}
	user.IsAdmin = isAdmin == 1

	rows, err := db.DB.Query(db.Rebind(`
		SELECT id, type, amount, description, created_at
		FROM transactions
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 50`), userID,
	)
	if err != nil {
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	transactions := []models.Transaction{}
	for rows.Next() {
		var tx models.Transaction
		err := rows.Scan(&tx.ID, &tx.Type, &tx.Amount, &tx.Description, &tx.CreatedAt)
		if err != nil {
			continue
		}
		tx.UserID = userID
		transactions = append(transactions, tx)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":         user,
		"transactions": transactions,
	})
}

func BorrowCoins(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var lastBorrow sql.NullString
	var balance, borrowed float64

	err := db.DB.QueryRow(db.Rebind(`
		SELECT balance, borrowed, last_borrow FROM users WHERE id = ?`), userID,
	).Scan(&balance, &borrowed, &lastBorrow)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	now := time.Now().UTC()
	today := now.Format("2006-01-02")

	if lastBorrow.Valid {
		lastBorrowDate := lastBorrow.String[:10]
		if lastBorrowDate == today {
			http.Error(w, "Daily loan already used today. Resets at midnight UTC.", http.StatusBadRequest)
			return
		}
	}

	borrowAmount := 400.0
	nowStr := now.Format(time.RFC3339)

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(db.Rebind(`
		UPDATE users
		SET balance = balance + ?, borrowed = borrowed + ?, last_borrow = ?
		WHERE id = ?`), borrowAmount, borrowAmount, nowStr, userID,
	)
	if err != nil {
		http.Error(w, "Failed to update balance", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(db.Rebind(`
		INSERT INTO transactions (id, user_id, type, amount, description, created_at)
		VALUES (?, ?, 'borrow', ?, 'Daily house loan', ?)`),
		uuid.New().String(), userID, borrowAmount, nowStr,
	)
	if err != nil {
		http.Error(w, "Failed to record transaction", http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"new_balance":    balance + borrowAmount,
		"total_borrowed": borrowed + borrowAmount,
		"message":        "400 coins borrowed from the house",
	})
}
