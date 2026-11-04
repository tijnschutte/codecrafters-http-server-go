package main

import (
	"errors"
	"io"
	"strings"
)

type Method string

const (
	GET  Method = "GET"
	POST Method = "POST"
)

type Request struct {
	method  Method
	path    string
	version string
	headers map[string]string
	body    string
}

func readRequest(conn io.Reader) (Request, error) {
	parsed := Request{}

	// Read data
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return parsed, err
	}
	msg := string(buf[:n])

	left, right, found := strings.Cut(msg, CRLF)
	if !found {
		return parsed, errors.New("invalid HTTP")
	}
	requestLine := strings.Split(left, " ")
	if len(requestLine) != 3 {
		return parsed, errors.New("request line should be exactly 3 parts")
	}
	parsed.method = Method(requestLine[0])
	parsed.path = requestLine[1]
	parsed.version = requestLine[2]

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
