package main

import (
	"log"

	"probubbly/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	if err := db.Init(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	_, err := db.DB.Exec(db.Rebind("UPDATE users SET is_admin = 1 WHERE login_id = ?"), "A0000")
	if err != nil {
		log.Fatal("Failed to update user:", err)
	}

	log.Println("Admin user A0000 updated successfully")
}
