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

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.EventDate == "" {
		http.Error(w, "Title and event date are required", http.StatusBadRequest)
		return
	}

	var username string
	err := db.DB.QueryRow(db.Rebind("SELECT username FROM users WHERE id = ?"), userID).Scan(&username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	eventID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err = db.DB.Exec(db.Rebind(`
		INSERT INTO events (id, title, description, event_date, event_time, creator_id, creator_name, status, yes_coins, no_coins, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'open', 0, 0, ?)`),
		eventID, req.Title, req.Description, req.EventDate, req.EventTime, userID, username, now,
	)
	if err != nil {
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	event := models.Event{
		ID:          eventID,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   req.EventDate,
		EventTime:   req.EventTime,
		CreatorID:   userID,
		CreatorName: username,
		Status:      "open",
		YesCoins:    0,
		NoCoins:     0,
		CreatedAt:   now,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func ListEvents(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	query := `
		SELECT e.id, e.title, e.description, e.event_date, e.event_time,
		       e.creator_id, u.username, e.status, e.yes_coins, e.no_coins,
		       e.outcome, e.resolved_at, e.created_at
		FROM events e
		JOIN users u ON e.creator_id = u.id`

	args := []interface{}{}
	if status != "" {
		query += " WHERE e.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY e.created_at DESC"

	rows, err := db.DB.Query(db.Rebind(query), args...)
	if err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []models.Event{}
	for rows.Next() {
		var e models.Event
		var eventTime, outcome, resolvedAt sql.NullString

		err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.EventDate, &eventTime,
			&e.CreatorID, &e.CreatorName, &e.Status, &e.YesCoins, &e.NoCoins,
			&outcome, &resolvedAt, &e.CreatedAt,
		)
		if err != nil {
			continue
		}

		if eventTime.Valid {
			e.EventTime = eventTime.String
		}
		if outcome.Valid {
			e.Outcome = outcome.String
		}
		if resolvedAt.Valid {
			e.ResolvedAt = resolvedAt.String
		}

		events = append(events, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func GetEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")

	var e models.Event
	var eventTime, outcome, resolvedAt sql.NullString

	err := db.DB.QueryRow(db.Rebind(`
		SELECT e.id, e.title, e.description, e.event_date, e.event_time,
		       e.creator_id, u.username, e.status, e.yes_coins, e.no_coins,
		       e.outcome, e.resolved_at, e.created_at
		FROM events e
		JOIN users u ON e.creator_id = u.id
		WHERE e.id = ?`), eventID,
	).Scan(
		&e.ID, &e.Title, &e.Description, &e.EventDate, &eventTime,
		&e.CreatorID, &e.CreatorName, &e.Status, &e.YesCoins, &e.NoCoins,
		&outcome, &resolvedAt, &e.CreatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
		return
	}

	if eventTime.Valid {
		e.EventTime = eventTime.String
	}
	if outcome.Valid {
		e.Outcome = outcome.String
	}
	if resolvedAt.Valid {
		e.ResolvedAt = resolvedAt.String
	}

	ratios := ratio.Compute(e.YesCoins, e.NoCoins)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"event":  e,
		"ratios": ratios,
	})
}
