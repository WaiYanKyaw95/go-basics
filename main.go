package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var statusText = map[int]string{
	200: "OK",
	201: "Created",
	400: "Bad Request",
	401: "Unauthorized",
	404: "Not Found",
	409: "Conflict",
	500: "Internal Server Error",
}

func main() {
	host := "localhost"
	port := "8080"

	// set the address
	address := host + ":" + port

	// initiate the connection with database.
	db := initDB()
	defer db.Close()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("server is listening on", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, db)
	}
}

func parseRequest(rawRequest string) (string, string, string) {
	var body string

	// separate headers and body, if available
	parts := strings.SplitN(rawRequest, "\r\n\r\n", 2)
	if len(parts) > 1 {
		body = parts[1]
	}
	headers := strings.Split(parts[0], "\r\n")

	// extract the first line that contains method, path and protocol
	firstLine := strings.Split(headers[0], " ")
	method := firstLine[0]
	path := firstLine[1]
	return method, path, body
}

func handleConnection(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	rawRequest := string(buffer[:n])
	method, path, body := parseRequest(rawRequest)
	// to check the length of path
	parts := strings.Split(path, "/")

	if method == "GET" && path == "/" {
		writeText(conn, 200, "Hello from the server side.")
	} else if method == "GET" && path == "/items" {
		rows, err := db.Query("SELECT * FROM items")
		if err != nil {
			writeText(conn, 500, "Database error.")
			return
		}
		defer rows.Close()
		var result = []Item{}
		for rows.Next() {
			var item Item
			rows.Scan(&item.ID, &item.Name)
			result = append(result, item)
		}
		if err := rows.Err(); err != nil {
			writeText(conn, 500, "Database error.")
			return
		}
		writeJSON(conn, 200, result)
	} else if method == "GET" && len(parts) == 3 && parts[1] == "items" {
		// catch bad request such as /items/abc
		id, err := strconv.Atoi(parts[2])
		if err != nil {
			writeText(conn, 400, "Bad Request.")
			return
		}
		row := db.QueryRow("SELECT * FROM items WHERE id = ?", id)
		var item Item
		err = row.Scan(&item.ID, &item.Name)
		if err == sql.ErrNoRows {
			writeText(conn, 404, "Item not found.")
			return
		} else if err != nil {
			writeText(conn, 500, "Database error.")
			return
		}
		writeJSON(conn, 200, item)
	} else if method == "POST" && path == "/items" {
		var item Item
		json.Unmarshal([]byte(body), &item)

		result, err := db.Exec("INSERT INTO items (name) VALUES (?)", item.Name)
		if err != nil {
			writeText(conn, 500, "Database error.")
			return
		}
		id, _ := result.LastInsertId()
		newItem := Item{ID: int(id), Name: item.Name}
		writeJSON(conn, 201, newItem)
	} else if method == "PUT" && len(parts) == 3 && parts[1] == "items" {
		var updatedBody Item
		// catch bad request such as /items/abc
		id, err := strconv.Atoi(parts[2])
		if err != nil {
			writeText(conn, 400, "Request could not be resolved.")
			return
		}
		json.Unmarshal([]byte(body), &updatedBody)

		var item Item
		row := db.QueryRow("SELECT * FROM items WHERE id = ?", id)
		err = row.Scan(&item.ID, &item.Name)
		if err == sql.ErrNoRows {
			writeText(conn, 404, "Item not found.")
			return
		} else if err != nil {
			writeText(conn, 500, "Database error.")
			return
		}

		_, err = db.Exec("UPDATE items SET name = ? WHERE id = ?", updatedBody.Name, id)
		if err != nil {
			writeText(conn, 500, "Database error.")
			return
		}

		updatedItem := Item{ID: id, Name: updatedBody.Name}
		writeJSON(conn, 200, updatedItem)
	} else if method == "DELETE" && len(parts) == 3 && parts[1] == "items" {
		// catch bad request such as /items/abc
		id, err := strconv.Atoi(parts[2])
		if err != nil {
			writeText(conn, 400, "Request could not be resolved.")
			return
		}

		row := db.QueryRow("SELECT * FROM items WHERE id = ?", id)
		// temporary Item to give it back to the client in the response.
		var deletedItem Item
		err = row.Scan(&deletedItem.ID, &deletedItem.Name)
		if err == sql.ErrNoRows {
			writeText(conn, 404, "Item not found.")
			return
		} else if err != nil {
			writeText(conn, 500, "Database error.")
			return
		}
		_, err = db.Exec("DELETE FROM items WHERE id = ?", id)
		if err != nil {
			writeText(conn, 500, "Database error.")
			return
		}
		writeJSON(conn, 200, deletedItem)
	} else if method == "POST" && path == "/register" {
		register(conn, body, db)
	} else if method == "POST" && path == "/login" {
		login(conn, body, db)
	} else if method == "POST" && path == "/logout" {
		logout(conn, rawRequest, db)
	} else {
		writeText(conn, 404, "Requested page not found.")
	}
}

func writeJSON(conn net.Conn, status int, data interface{}) {
	body, _ := json.Marshal(data)
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: application/json\r\n\r\n%s", status, statusText[status], string(body))
	conn.Write([]byte(response))
}

func writeText(conn net.Conn, status int, message string) {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: text/html\r\n\r\n%s", status, statusText[status], message)
	conn.Write([]byte(response))
}
