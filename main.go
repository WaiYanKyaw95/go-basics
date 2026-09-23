package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var items []Item
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

	fmt.Println(body)

	if method == "GET" && path == "/" {
		response = "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/html\r\n" +
			"\r\n" +
			"Hello from the server side."
	} else {
		response = "HTTP/1.1 404 NOT FOUND\r\n" +
			"Content-Type: text/html\r\n" +
			"\r\n" +
			"Not Found."
	}

	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("error writing response:", err)
	}
}
