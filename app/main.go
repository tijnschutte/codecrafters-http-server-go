package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const CRLF = "\r\n"

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer listener.Close()

	for {
		// Block until we receive an incoming connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	// Ensure we close the connection after we're done
	defer conn.Close()

	req, err := readRequest(conn)
	if err != nil {
		return
	}
	switch {
	case strings.HasPrefix(req.path, "/echo/"):
		body, _ := strings.CutPrefix(req.path, "/echo/")
		Response{
			status: "HTTP/1.1 200 OK",
			headers: map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": strconv.Itoa(len(body)),
			},
			body: body,
		}.Write(conn)
	case req.path == "/user-agent":
		body := req.headers["user-agent"]
		Response{
			status: "HTTP/1.1 200 OK",
			headers: map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": strconv.Itoa(len(body)),
			},
			body: body,
		}.Write(conn)
	case req.path == "/":
		Response{status: "HTTP/1.1 200 OK"}.Write(conn)
	default:
		Response{status: "HTTP/1.1 404 Not Found"}.Write(conn)
	}
}
