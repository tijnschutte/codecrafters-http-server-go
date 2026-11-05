package main

import (
	"bufio"
	"errors"
	"io"
	"strconv"
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

// r must live as long as the connection: bufio reads ahead into the next request.
func readRequest(r *bufio.Reader) (Request, error) {
	readline := func() (string, error) {
		line, err := r.ReadString('\n')
		return strings.TrimSuffix(line, "\r\n"), err
	}

	line, err := readline()
	if err != nil {
		return Request{}, err
	}
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return Request{}, errors.New("request line should be exactly 3 parts")
	}
	req := Request{
		method:  Method(parts[0]),
		path:    parts[1],
		version: parts[2],
		headers: map[string]string{},
	}
	for {
		header, err := readline()
		if err != nil {
			return Request{}, err
		}
		if header == "" {
			break
		}
		k, v, _ := strings.Cut(header, ":")
		req.headers[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	n := 0
	if cl, ok := req.headers["content-length"]; ok {
		n, err = strconv.Atoi(cl)
		if err != nil || n < 0 || n > 10<<20 {
			return Request{}, errors.New("bad content-length")
		}
	}
	body := make([]byte, n)
	if _, err := io.ReadFull(r, body); err != nil {
		return Request{}, err
	}
	req.body = string(body)

	return req, nil
}
