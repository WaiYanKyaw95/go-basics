package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "ledserver.db")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
						id INTEGER PRIMARY KEY AUTOINCREMENT,
						name TEXT NOT NULL
			)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
