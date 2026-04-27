package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"probubbly/internal/auth"
	"probubbly/internal/db"
	"probubbly/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	if err := db.Init(); err != nil {
		log.Fatal("Database initialisation failed:", err)
	}

	if err := db.ApplySchema(); err != nil {
		log.Fatal("Schema application failed:", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"https://probubbly-app.pages.dev",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Probubbly API is running")
	})

	// Public routes with strict rate limiting
	r.Group(func(r chi.Router) {
		// 10 requests per minute per IP for auth endpoints
		r.Use(httprate.LimitByIP(10, time.Minute))
		r.Post("/api/auth/signup", handlers.Signup)
		r.Post("/api/auth/login", handlers.Login)
	})

	// Protected routes with generous rate limiting
	r.Group(func(r chi.Router) {
		// 100 requests per minute per IP for normal usage
		r.Use(httprate.LimitByIP(100, time.Minute))
		r.Use(auth.AuthMiddleware)
		r.Get("/api/events", handlers.ListEvents)
		r.Post("/api/events", handlers.CreateEvent)
		r.Get("/api/events/{id}", handlers.GetEvent)
		r.Post("/api/events/{id}/predict", handlers.PlacePrediction)
		r.Get("/api/wallet", handlers.GetWallet)
		r.Post("/api/wallet/borrow", handlers.BorrowCoins)
		r.Post("/api/events/{id}/resolve", handlers.ResolveEvent)
	})

	// Admin routes
	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(30, time.Minute))
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
