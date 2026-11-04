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
	defer conn.Close()
	persistence := true
	for persistence {
		req, err := readRequest(conn)
		if err != nil {
			Response{status: Bad}.Write(conn)
			continue
		}
		persistence = req.headers["Connection"] != "close"

		switch {
		case strings.HasPrefix(req.path, "/files/"):
			filename, _ := strings.CutPrefix(req.path, "/files/")
			if !filepath.IsLocal(filename) {
				Response{status: Bad}.Write(conn)
				continue
			}
			path := filepath.Join(dir, filename)
			switch req.method {
			case GET:
				data, err := os.ReadFile(path)
				if err != nil {
					Response{status: NotFound}.Write(conn)
					continue
				}
				Response{
					status: Ok,
					headers: map[string]string{
						"Content-Type": "application/octet-stream",
					},
					body: string(data),
				}.Write(conn)
			case POST:
				err := os.WriteFile(path, []byte(req.body), 0o644)
				if err != nil {
					Response{status: InternalServerError}.Write(conn)
					continue
				}
				Response{status: Created}.Write(conn)
			}
		case strings.HasPrefix(req.path, "/echo/"):
			body, _ := strings.CutPrefix(req.path, "/echo/")
			Response{
				status: Ok,
				headers: map[string]string{
					"Content-Type": "text/plain",
				},
				body: body,
			}.Write(conn)
		case req.path == "/user-agent":
			body := req.headers["user-agent"]
			Response{
				status: Ok,
				headers: map[string]string{
					"Content-Type": "text/plain",
				},
				body: body,
			}.Write(conn)
		case req.path == "/":
			Response{status: Ok}.Write(conn)
		default:
			Response{status: NotFound}.Write(conn)
		}
	}
}
