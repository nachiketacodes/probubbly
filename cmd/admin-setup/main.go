package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "probubbly.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("UPDATE users SET is_admin = 1 WHERE login_id = 'A0000'")
	if err != nil {
		log.Fatal("Failed to update user:", err)
	}

	log.Println("Admin user A0000 updated successfully")
}