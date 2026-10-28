package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

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

type Request struct {
	method  string
	path    string
	version string
	headers map[string]string
	body    string
}

const CRLF = "\r\n"

func parseRequest(msg string) (Request, error) {
	// GET /echo/abc HTTP/1.1\r\nHost: localhost:4221\r\nUser-Agent: curl/7.64.1\r\nAccept: */*\r\n\r\n
	parsed := Request{}

	left, right, found := strings.Cut(msg, CRLF)
	if !found {
		return parsed, errors.New("invalid HTTP")
	}
	requestLine := strings.Split(left, " ")
	if len(requestLine) != 3 {
		return parsed, errors.New("request line should be exactly 3 parts")
	}
	parsed.method, parsed.path, parsed.version = requestLine[0], requestLine[1], requestLine[2]

	parsed.headers = map[string]string{}
	left, right, found = strings.Cut(right, CRLF)
	for found {
		key, value, keepGoing := strings.Cut(left, ": ")
		if !keepGoing {
			break
		}
		parsed.headers[strings.ToLower(key)] = value
		left, right, found = strings.Cut(right, CRLF)
	}

	body, _, _ := strings.Cut(right, CRLF)
	parsed.body = body

	return parsed, nil
}

type Response struct {
	status  string
	headers map[string]string
	body    string
}

func (r Response) Write(w io.Writer) error {
	bw := bufio.NewWriter(w)

	bw.WriteString(r.status)
	bw.WriteString(CRLF)
	for k, v := range r.headers {
		bw.WriteString(k)
		bw.WriteString(": ")
		bw.WriteString(v)
		bw.WriteString(CRLF)
	}
	bw.WriteString(CRLF)
	bw.WriteString(r.body)
	err := bw.Flush()
	return err
}

func handleClient(conn net.Conn) {
	// Ensure we close the connection after we're done
	defer conn.Close()

	// Read data
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	req, _ := parseRequest(string(buf[:n]))
	switch {
	case strings.HasPrefix(req.path, "/user-agent/"):
		body := req.headers["user-agent"]
		Response{
			status: "HTTP/1.1 200 OK",
			headers: map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": strconv.Itoa(len(body)),
			},
			body: body,
		}.Write(conn)
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
	case req.path == "/":
		Response{status: "HTTP/1.1 200 OK"}.Write(conn)
	default:
		Response{status: "HTTP/1.1 404 Not Found\r\n\r\n"}.Write(conn)
	}
}
