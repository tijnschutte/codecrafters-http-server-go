package main

import (
	"bufio"
	"io"
	"strconv"
)

type STATUS string

const (
	NotFound            STATUS = "HTTP/1.1 404 Not Found"
	Ok                  STATUS = "HTTP/1.1 200 OK"
	Created             STATUS = "HTTP/1.1 201 Created"
	Bad                 STATUS = "HTTP/1.1 400 Bad Request"
	InternalServerError STATUS = "HTTP/1.1 500 Internal Server Error"
)

type Response struct {
	status  STATUS
	headers map[string]string
	body    string
}

func (r Response) Write(w io.Writer, close bool) error {
	bw := bufio.NewWriter(w)

	bw.WriteString(string(r.status))
	bw.WriteString(CRLF)
	for k, v := range r.headers {
		bw.WriteString(k)
		bw.WriteString(": ")
		bw.WriteString(v)
		bw.WriteString(CRLF)
	}
	if close {
		bw.WriteString("Connection: close")
		bw.WriteString(CRLF)
	}
	bw.WriteString("Content-Length: ")
	bw.WriteString(strconv.Itoa(len(r.body)))
	bw.WriteString(CRLF)
	bw.WriteString(CRLF)
	bw.WriteString(r.body)
	err := bw.Flush()
	return err
}
