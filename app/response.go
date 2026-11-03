package main

import (
	"bufio"
	"io"
)

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
