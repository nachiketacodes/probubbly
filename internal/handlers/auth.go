package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"time"

	"probubbly/internal/db"
	"probubbly/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	matched, _ := regexp.MatchString(`^[A-Za-z][0-9]{4}$`, req.LoginID)
	if !matched {
		http.Error(w, "Login ID must be 1 letter followed by 4 digits (e.g. A1234)", http.StatusBadRequest)
		return
	}

	pinMatched, _ := regexp.MatchString(`^[0-9]{4}$`, req.PIN)
	if !pinMatched {
		http.Error(w, "PIN must be exactly 4 digits", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
    http.Error(w, "Username is required", http.StatusBadRequest)
    return
}
if len(req.Username) > 30 {
    http.Error(w, "Username must be 30 characters or less", http.StatusBadRequest)
    return
}

	var existing string
	err := db.DB.QueryRow(db.Rebind("SELECT id FROM users WHERE login_id = ?"), req.LoginID).Scan(&existing)
	if err != sql.ErrNoRows {
		http.Error(w, "Login ID already taken", http.StatusConflict)
		return
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	userID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err = db.DB.Exec(db.Rebind(`
		INSERT INTO users (id, login_id, pin_hash, username, balance, borrowed, is_admin, joined_at)
		VALUES (?, ?, ?, ?, 500.0, 0.0, 0, ?)`),
		userID, req.LoginID, string(pinHash), req.Username, now,
	)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec(db.Rebind(`
		INSERT INTO transactions (id, user_id, type, amount, description, created_at)
		VALUES (?, ?, 'signup', 500.0, 'Welcome bonus', ?)`),
		uuid.New().String(), userID, now,
	)
	if err != nil {
		http.Error(w, "Failed to record transaction", http.StatusInternalServerError)
		return
	}

	token, err := generateToken(userID, false)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:       userID,
		Username: req.Username,
		Balance:  500.0,
		IsAdmin:  false,
		JoinedAt: now,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, User: user})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var user models.User
	var pinHash string
	var lastBorrow sql.NullString
	var isAdmin int

	err := db.DB.QueryRow(db.Rebind(`
		SELECT id, login_id, pin_hash, username, balance, borrowed, last_borrow, is_admin, joined_at
		FROM users WHERE login_id = ?`), req.LoginID,
	).Scan(
		&user.ID, &user.LoginID, &pinHash, &user.Username,
		&user.Balance, &user.Borrowed, &lastBorrow,
		&isAdmin, &user.JoinedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Invalid Login ID or PIN", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if lastBorrow.Valid {
		user.LastBorrow = lastBorrow.String
	} else {
		user.LastBorrow = ""
	}
	user.IsAdmin = isAdmin == 1

	if err := bcrypt.CompareHashAndPassword([]byte(pinHash), []byte(req.PIN)); err != nil {
		http.Error(w, "Invalid Login ID or PIN", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.ID, user.IsAdmin)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{Token: token, User: user})
}

func generateToken(userID string, isAdmin bool) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"user_id":  userID,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
