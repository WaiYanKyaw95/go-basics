package main

import (
	"database/sql"
	"encoding/json"
	"net"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

func register(conn net.Conn, body string, db *sql.DB) {
	var info User
	json.Unmarshal([]byte(body), &info)

	hash, err := bcrypt.GenerateFromPassword([]byte(info.Password), bcrypt.DefaultCost)
	if err != nil {
		writeText(conn, 500, "Internal Server Error", "Could not hash password.")
		return
	}
	result, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", info.Username, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeText(conn, 409, "Conflict", "Username already taken.")
			return
		}
		writeText(conn, 500, "Internal Server Error", "Database error.")
		return
	}
	id, _ := result.LastInsertId()
	newUser := User{ID: int(id), Username: info.Username}
	writeJSON(conn, 201, "Created", newUser)
}

func login(conn net.Conn, body string, db *sql.DB) {
	// get the username and password
	// check if the user is there
	// if not found, 401
	// compare the password with hash
	// if not match, 401
	// generate a token, store in sessions and send cookie
}
