package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"probubbly/internal/auth"
	"probubbly/internal/db"
	"probubbly/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialise database connection
	if err := db.Init(); err != nil {
		log.Fatal("Database initialisation failed:", err)
	}

	// Apply database schema
	if err := db.ApplySchema(); err != nil {
		log.Fatal("Schema application failed:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://*.pages.dev"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Health check
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Probubbly API is running")
	})

	// Public routes — no authentication required
	r.Post("/api/auth/signup", handlers.Signup)
	r.Post("/api/auth/login", handlers.Login)

	// Protected routes — authentication required
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)

		// Events
		r.Get("/api/events", handlers.ListEvents)
		r.Post("/api/events", handlers.CreateEvent)
		r.Get("/api/events/{id}", handlers.GetEvent)
		r.Post("/api/events/{id}/predict", handlers.PlacePrediction)

		// Wallet
		r.Get("/api/wallet", handlers.GetWallet)
		r.Post("/api/wallet/borrow", handlers.BorrowCoins)

		// Resolution (admin or creator only — checked inside handler)
		r.Post("/api/events/{id}/resolve", handlers.ResolveEvent)
	})

	// Admin-only routes
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Use(auth.AdminMiddleware)

		r.Get("/api/admin/stats", handlers.GetPlatformStats)
		r.Get("/api/admin/users", handlers.ListAllUsers)
		r.Get("/api/admin/user", handlers.GetUserDetail)
		r.Get("/api/admin/house-ledger", handlers.GetHouseLedger)
	})

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}