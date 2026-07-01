package http

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

type Response struct {
	Version    string
	StatusCode int
	Status     string
	Headers    map[string]string
}

func ReadResponse(conn net.Conn) (*Response, error) {

	reader := bufio.NewReader(conn)

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimRight(line, "\r\n")

	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid HTTP response")
	}

	code, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	tp := textproto.NewReader(reader)

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)

	for k, v := range mimeHeader {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &Response{
		Version:    parts[0],
		StatusCode: code,
		Status:     parts[2],
		Headers:    headers,
	}, nil
}