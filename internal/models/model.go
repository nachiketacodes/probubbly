package models

type User struct {
	ID         string  `json:"id"`
	LoginID    string  `json:"login_id"`
	Username   string  `json:"username"`
	Balance    float64 `json:"balance"`
	Borrowed   float64 `json:"borrowed"`
	LastBorrow string  `json:"last_borrow"`
	IsAdmin    bool    `json:"is_admin"`
	JoinedAt   string  `json:"joined_at"`
}

type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
	EventTime   string `json:"event_time"`
	CreatorID   string `json:"creator_id"`
	CreatorName string `json:"creator_name"`
	Status      string `json:"status"`
	YesCoins    int    `json:"yes_coins"`
	NoCoins     int    `json:"no_coins"`
	Outcome     string `json:"outcome"`
	ResolvedAt  string `json:"resolved_at"`
	CreatedAt   string `json:"created_at"`
}

type Prediction struct {
	ID        string  `json:"id"`
	EventID   string  `json:"event_id"`
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Side      string  `json:"side"`
	Amount    int     `json:"amount"`
	Ratio     float64 `json:"ratio"`
	Payout    float64 `json:"payout"`
	CreatedAt string  `json:"created_at"`
}

type Transaction struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
}

type SignupRequest struct {
	LoginID  string `json:"login_id"`
	PIN      string `json:"pin"`
	Username string `json:"username"`
}

type LoginRequest struct {
	LoginID string `json:"login_id"`
	PIN     string `json:"pin"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type PredictRequest struct {
	Side   string `json:"side"`
	Amount int    `json:"amount"`
}

type CreateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
	EventTime   string `json:"event_time"`
}

type ResolveRequest struct {
	Outcome string `json:"outcome"`
}

type RatioResult struct {
	Yes      float64 `json:"yes"`
	No       float64 `json:"no"`
	YesPct   int     `json:"yes_pct"`
	NoPct    int     `json:"no_pct"`
	HouseCut float64 `json:"house_cut"`
}
