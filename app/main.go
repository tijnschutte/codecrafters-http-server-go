package main

import (
	"bufio"
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
	close := false
	r := bufio.NewReader(conn)
	for !close {
		req, err := readRequest(r)
		if err != nil {
			// EOF or broken conn: the client is gone, stop
			return
		}
		close = req.headers["connection"] == "close"

		fmt.Println(req)
		fmt.Println(close)

		switch {
		case strings.HasPrefix(req.path, "/files/"):
			filename, _ := strings.CutPrefix(req.path, "/files/")
			if !filepath.IsLocal(filename) {
				Response{status: Bad}.Write(conn, close)
				continue
			}
			path := filepath.Join(dir, filename)
			switch req.method {
			case GET:
				data, err := os.ReadFile(path)
				if err != nil {
					Response{status: NotFound}.Write(conn, close)
					continue
				}
				Response{
					status: Ok,
					headers: map[string]string{
						"Content-Type": "application/octet-stream",
					},
					body: string(data),
				}.Write(conn, close)
			case POST:
				err := os.WriteFile(path, []byte(req.body), 0o644)
				if err != nil {
					Response{status: InternalServerError}.Write(conn, close)
					continue
				}
				Response{status: Created}.Write(conn, close)
			}
		case strings.HasPrefix(req.path, "/echo/"):
			body, _ := strings.CutPrefix(req.path, "/echo/")
			Response{
				status: Ok,
				headers: map[string]string{
					"Content-Type": "text/plain",
				},
				body: body,
			}.Write(conn, close)
		case req.path == "/user-agent":
			body := req.headers["user-agent"]
			Response{
				status: Ok,
				headers: map[string]string{
					"Content-Type": "text/plain",
				},
				body: body,
			}.Write(conn, close)
		case req.path == "/":
			Response{status: Ok}.Write(conn, close)
		default:
			Response{status: NotFound}.Write(conn, close)
		}
	}
}
