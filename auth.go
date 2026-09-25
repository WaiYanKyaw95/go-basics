package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func register(conn net.Conn, body string, db *sql.DB) {
	var info LoginInput
	json.Unmarshal([]byte(body), &info)

	hash, err := bcrypt.GenerateFromPassword([]byte(info.Password), bcrypt.DefaultCost)
	if err != nil {
		writeText(conn, 500, "Could not hash password.")
		return
	}
	result, err := db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", info.Username, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeText(conn, 409, "Username already taken.")
			return
		}
		writeText(conn, 500, "Database error.")
		return
	}
	id, _ := result.LastInsertId()
	newUser := User{ID: int(id), Username: info.Username}
	writeJSON(conn, 201, newUser)
}

func login(conn net.Conn, body string, db *sql.DB) {
	// get the username and password
	var info LoginInput // what the client sent
	var user User       // what we retrieved from the database
	json.Unmarshal([]byte(body), &info)

	// check if the user is there
	row := db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", info.Username)
	err := row.Scan(&user.ID, &user.Username, &user.Password)

	// if not found, 401
	if err == sql.ErrNoRows {
		writeText(conn, 401, "Invalid Crendentials.")
		return
	} else if err != nil {
		writeText(conn, 500, "Database error.")
		return
	}

	// compare the password with hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(info.Password))

	// if not match, 401
	if err != nil {
		writeText(conn, 401, "Invalid Crendentials.")
		return
	}
	// generate a token, store in sessions and send cookie
	bytes := make([]byte, 32)
	rand.Read(bytes)
	token := hex.EncodeToString(bytes)

	_, err = db.Exec("INSERT INTO sessions (user_id, token, created_at) VALUES (?, ?, ?)", user.ID, token, time.Now().Format(time.RFC3339))
	if err != nil {
		writeText(conn, 500, "Database error.")
		return
	}

	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nSet-Cookie: session=%s; HttpOnly\r\nContent-Type: application/json\r\n\r\n%s",
		token,
		`{"message": "logged in successfully"}`,
	)
	conn.Write([]byte(response))
}

func logout(conn net.Conn, rawRequest string, db *sql.DB) {
	// get the session token from the request
	token := getCookie(rawRequest, "session")
	// if no token -> 401
	if token == "" {
		writeText(conn, 401, "No token included")
		return
	}
	// DELETE FROM sessions WHERE token = ?
	_, err := db.Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		writeText(conn, 500, "Database error.")
		return
	}
	// return 200 "logged out successfully"
	writeText(conn, 200, "logged out successfully")
}

func getCookie(rawRequest string, name string) string {
	// find the cookie by name
	headers := strings.Split(rawRequest, "\r\n")
	for _, header := range headers {
		if strings.HasPrefix(header, "Cookie:") {
			if strings.Contains(header, name) {
				sessionKey := strings.SplitN(header, "=", 2)
				value := strings.SplitN(sessionKey[1], ";", 2)[0]
				// return its value
				return strings.TrimSpace(value)
			}
		}
	}
	// return "" if not found
	return ""
}

func getSessionUser(rawRequest string, db *sql.DB) (*User, error) {
	var user User

	// get the session cookie
	token := getCookie(rawRequest, "session")

	// if no cookie -> return error
	if token == "" {
		return nil, fmt.Errorf("unauthorized")
	}

	// look up token in the sessions table
	row := db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", token)
	err := row.Scan(&user.ID)

	// if no session -> token invalid -> return error
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("unauthorized")
	} else if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	// get the actual user
	row = db.QueryRow("SELECT id, username FROM users WHERE id = ?", user.ID)
	err = row.Scan(&user.ID, &user.Username)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("unauthorized")
	} else if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}
	// return the user
	return &user, nil
}
