-- Users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    login_id TEXT UNIQUE NOT NULL,
    pin_hash TEXT NOT NULL,
    username TEXT NOT NULL,
    balance INTEGER NOT NULL DEFAULT 500,
    borrowed INTEGER NOT NULL DEFAULT 0,
    last_borrow TEXT,
    is_admin INTEGER NOT NULL DEFAULT 0,
    joined_at TEXT NOT NULL
);

-- Events table
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    event_date TEXT NOT NULL,
    event_time TEXT,
    creator_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    yes_coins INTEGER NOT NULL DEFAULT 0,
    no_coins INTEGER NOT NULL DEFAULT 0,
    outcome TEXT,
    resolved_at TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (creator_id) REFERENCES users(id)
);

-- Predictions table
CREATE TABLE IF NOT EXISTS predictions (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    side TEXT NOT NULL,
    amount INTEGER NOT NULL,
    ratio REAL NOT NULL,
    payout INTEGER,
    created_at TEXT NOT NULL,
    FOREIGN KEY (event_id) REFERENCES events(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    type TEXT NOT NULL,
    amount INTEGER NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Indices for faster queries
CREATE INDEX IF NOT EXISTS idx_predictions_event ON predictions(event_id);
CREATE INDEX IF NOT EXISTS idx_predictions_user ON predictions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
