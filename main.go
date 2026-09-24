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

var items = []Item{}
var nextID int = 1

func main() {
	host := "localhost"
	port := "8080"

	// set the address
	address := host + ":" + port

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
	var response string

	defer conn.Close()

	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	method, path, body := parseRequest(string(buffer[:n]))
	// to check the length of path
	parts := strings.Split(path, "/")

	if method == "GET" && path == "/" {
		message := "Hello from the server side."
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n%s", message)
	} else if method == "GET" && path == "/items" {
		rows, err := db.Query("SELECT * FROM items")
		if err != nil {
			message := "Database error."
			response = fmt.Sprintf("HTTP/1.1 500 Internal Server Error\r\n\r\n%s", message)
			conn.Write([]byte(response))
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
			message := "Database error."
			response = fmt.Sprintf("HTTP/1.1 500 Internal Server Error\r\n\r\n%s", message)
			conn.Write([]byte(response))
			return
		}
		message, _ := json.Marshal(result)
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n%s", string(message))
	} else if method == "GET" && len(parts) == 3 && parts[1] == "items" {
		// catch bad request such as /items/abc
		id, err := strconv.Atoi(parts[2])
		if err != nil {
			message := "Bad Request."
			response = fmt.Sprintf("HTTP/1.1 400 Bad Request\r\nContent-Type: text/html\r\n\r\n%s", message)
			conn.Write([]byte(response))
			return
		}
		row := db.QueryRow("SELECT * FROM items WHERE id = ?", id)
		var item Item
		err = row.Scan(&item.ID, &item.Name)
		if err == sql.ErrNoRows {
			message := "Item not found."
			response = fmt.Sprintf("HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\n\r\n%s", message)
			conn.Write([]byte(response))
			return
		} else if err != nil {
			message := "Internal Server Error."
			response = fmt.Sprintf("HTTP/1.1 500 Internal Server Error\r\nContent-Type: text/html\r\n\r\n%s", message)
			conn.Write([]byte(response))
			return
		}
		message, _ := json.Marshal(item)
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n%s", string(message))
	} else if method == "DELETE" && len(parts) == 3 && parts[1] == "items" {
		foundIndex := -1
		var deletedItem Item
		id, err := strconv.Atoi(parts[2])
		if err != nil {
			message := "Bad Request."
			response = fmt.Sprintf("HTTP/1.1 400 Bad Request\r\nContent-Type: text/html\r\n\r\n%s", message)
			conn.Write([]byte(response))
			return
		}
		for index, item := range items {
			if item.ID == id {
				foundIndex = index
				deletedItem = item
				break
			}
		}
		if foundIndex == -1 {
			message := "Item not found."
			response = fmt.Sprintf("HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\n\r\n%s", message)
		} else {
			// returns with 200 and deleted item
			items = append(items[:foundIndex], items[foundIndex+1:]...)
			message, _ := json.Marshal(deletedItem)
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n%s", string(message))
		}
	} else if method == "PUT" && len(parts) == 3 && parts[1] == "items" {
		var updatedItem Item
		foundIndex := -1
		id, err := strconv.Atoi(parts[2])
		if err != nil {
			message := "Bad Request."
			response = fmt.Sprintf("HTTP/1.1 400 Bad Request\r\nContent-Type: text/html\r\n\r\n%s", message)
			conn.Write([]byte(response))
			return
		}
		json.Unmarshal([]byte(body), &updatedItem)

		for index, item := range items {
			if item.ID == id {
				foundIndex = index
				break
			}
		}

		if foundIndex == -1 {
			message := "Item not found."
			response = fmt.Sprintf("HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\n\r\n%s", message)
		} else {
			// returns with 200 and updated item
			items[foundIndex].Name = updatedItem.Name
			message, _ := json.Marshal(items[foundIndex])
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n%s", string(message))
		}

	} else if method == "POST" && path == "/items" {
		var item Item
		json.Unmarshal([]byte(body), &item)

		result, err := db.Exec("INSERT INTO items (name) VALUES (?)", item.Name)
		if err != nil {
			message := "Database error."
			response = fmt.Sprintf("HTTP/1.1 500 Internal Server Error\r\n\r\n%s", message)
			conn.Write([]byte(response))
			return
		}
		id, _ := result.LastInsertId()
		newItem := Item{ID: int(id), Name: item.Name}
		message, _ := json.Marshal(newItem)
		response = fmt.Sprintf("HTTP/1.1 201 Created\r\nContent-Type: application/json\r\n\r\n%s", string(message))
	} else {
		message := "Not Found."
		response = fmt.Sprintf("HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\n\r\n%s", message)
	}

	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("error writing response:", err)
	}
}
