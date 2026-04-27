package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init() error {
	databaseURL := os.Getenv("DATABASE_URL")

	var err error
	if databaseURL != "" {
		DB, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return err
		}
	} else {
		DB, err = sql.Open("sqlite", "probubbly.db")
		if err != nil {
			return err
		}
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Database connected successfully")
	return nil
}

func ApplySchema() error {
	databaseURL := os.Getenv("DATABASE_URL")

	var schema string
	if databaseURL != "" {
		schema = postgresSchema
	} else {
		schema = sqliteSchema
	}

	_, err := DB.Exec(schema)
	if err != nil {
		return err
	}

	log.Println("Database schema applied successfully")
	return nil
}

// IsPostgres returns true if using PostgreSQL
func IsPostgres() bool {
	return os.Getenv("DATABASE_URL") != ""
}

// Rebind replaces ? placeholders with $1, $2, etc. for PostgreSQL
func Rebind(query string) string {
	if !IsPostgres() {
		return query
	}
	count := 0
	var result strings.Builder
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			count++
			result.WriteString(fmt.Sprintf("$%d", count))
		} else {
			result.WriteByte(query[i])
		}
	}
	return result.String()
}

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	login_id TEXT UNIQUE NOT NULL,
	pin_hash TEXT NOT NULL,
	username TEXT NOT NULL,
	balance REAL NOT NULL DEFAULT 500,
	borrowed REAL NOT NULL DEFAULT 0,
	last_borrow TEXT,
	is_admin INTEGER NOT NULL DEFAULT 0,
	joined_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT,
	event_date TEXT NOT NULL,
	event_time TEXT,
	creator_id TEXT NOT NULL,
	creator_name TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'open',
	yes_coins INTEGER NOT NULL DEFAULT 0,
	no_coins INTEGER NOT NULL DEFAULT 0,
	outcome TEXT,
	resolved_at TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (creator_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS predictions (
	id TEXT PRIMARY KEY,
	event_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	side TEXT NOT NULL,
	amount INTEGER NOT NULL,
	ratio REAL NOT NULL,
	payout REAL,
	created_at TEXT NOT NULL,
	FOREIGN KEY (event_id) REFERENCES events(id),
	FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS transactions (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	type TEXT NOT NULL,
	amount REAL NOT NULL,
	description TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS house_ledger (
	id TEXT PRIMARY KEY,
	event_id TEXT NOT NULL,
	prediction_id TEXT NOT NULL,
	cut_amount REAL NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY (event_id) REFERENCES events(id)
);

CREATE INDEX IF NOT EXISTS idx_predictions_event_id ON predictions(event_id);
CREATE INDEX IF NOT EXISTS idx_predictions_user_id ON predictions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
`

const postgresSchema = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	login_id TEXT UNIQUE NOT NULL,
	pin_hash TEXT NOT NULL,
	username TEXT NOT NULL,
	balance DOUBLE PRECISION NOT NULL DEFAULT 500,
	borrowed DOUBLE PRECISION NOT NULL DEFAULT 0,
	last_borrow TEXT,
	is_admin INTEGER NOT NULL DEFAULT 0,
	joined_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT,
	event_date TEXT NOT NULL,
	event_time TEXT,
	creator_id TEXT NOT NULL,
	creator_name TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'open',
	yes_coins INTEGER NOT NULL DEFAULT 0,
	no_coins INTEGER NOT NULL DEFAULT 0,
	outcome TEXT,
	resolved_at TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (creator_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS predictions (
	id TEXT PRIMARY KEY,
	event_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	side TEXT NOT NULL,
	amount INTEGER NOT NULL,
	ratio DOUBLE PRECISION NOT NULL,
	payout DOUBLE PRECISION,
	created_at TEXT NOT NULL,
	FOREIGN KEY (event_id) REFERENCES events(id),
	FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS transactions (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	type TEXT NOT NULL,
	amount DOUBLE PRECISION NOT NULL,
	description TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS house_ledger (
	id TEXT PRIMARY KEY,
	event_id TEXT NOT NULL,
	prediction_id TEXT NOT NULL,
	cut_amount DOUBLE PRECISION NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY (event_id) REFERENCES events(id)
);

CREATE INDEX IF NOT EXISTS idx_predictions_event_id ON predictions(event_id);
CREATE INDEX IF NOT EXISTS idx_predictions_user_id ON predictions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
`
