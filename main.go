package main

import (
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
		go handleConnection(conn)
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

func handleConnection(conn net.Conn) {
	var response string

	defer conn.Close()

	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	method, path, body := parseRequest(string(buffer[:n]))
	parts := strings.Split(path, "/")
	// fmt.Println(body)

	if method == "GET" && path == "/" {
		message := "Hello from the server side."
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n%s", message)
	} else if method == "GET" && path == "/items" {
		message, _ := json.Marshal(items)
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n%s", string(message))
	} else if method == "GET" && len(parts) == 3 && parts[1] == "items" {
		found := false
		id, err := strconv.Atoi(parts[2])
		if err != nil {
			message := "Bad Request."
			response = fmt.Sprintf("HTTP/1.1 400 Bad Request\r\nContent-Type: text/html\r\n\r\n%s", message)
			conn.Write([]byte(response))
			return
		}
		for _, item := range items {
			if item.ID == id {
				found = true
				message, _ := json.Marshal(item)
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n%s", string(message))
				break
			}
		}
		if !found {
			message := "Item not found."
			response = fmt.Sprintf("HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\n\r\n%s", message)
		}
	} else if method == "POST" && path == "/items" {
		var item Item
		json.Unmarshal([]byte(body), &item)

		item.ID = nextID
		nextID++

		items = append(items, item)
		message, _ := json.Marshal(item)
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
