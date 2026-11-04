package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const CRLF = "\r\n"

func main() {
	dir := flag.String("directory", "", "root dir for /files")
	flag.Parse()
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

		go handleClient(conn, *dir)
	}
}

func handleClient(conn net.Conn, dir string) {
	// Ensure we close the connection after we're done
	defer conn.Close()

	req, err := readRequest(conn)
	if err != nil {
		return
	}
	fmt.Println(req)

	switch {
	case strings.HasPrefix(req.path, "/files/"):
		switch req.method {
		case GET:
			filename, _ := strings.CutPrefix(req.path, "/files/")
			data, err := os.ReadFile(filepath.Join(dir, filename))
			if err != nil {
				Response{status: "HTTP/1.1 404 Not Found"}.Write(conn)
				return
			}
			Response{
				status: "HTTP/1.1 200 OK",
				headers: map[string]string{
					"Content-Type": "application/octet-stream",
				},
				body: string(data),
			}.Write(conn)
		case POST:
		}
	case strings.HasPrefix(req.path, "/echo/"):
		body, _ := strings.CutPrefix(req.path, "/echo/")
		Response{
			status: "HTTP/1.1 200 OK",
			headers: map[string]string{
				"Content-Type": "text/plain",
			},
			body: body,
		}.Write(conn)
	case req.path == "/user-agent":
		body := req.headers["user-agent"]
		Response{
			status: "HTTP/1.1 200 OK",
			headers: map[string]string{
				"Content-Type": "text/plain",
			},
			body: body,
		}.Write(conn)
	case req.path == "/":
		Response{status: "HTTP/1.1 200 OK"}.Write(conn)
	default:
		Response{status: "HTTP/1.1 404 Not Found"}.Write(conn)
	}
}
