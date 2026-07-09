package http

import (
	"bufio"
	"fmt"
	"net/textproto"
	"strconv"
	"strings"
)

type Response struct {
	Version       string
	StatusCode    int
	Status        string
	Headers       map[string]string
	ContentLength int64
	Chunked       bool
}

func ReadResponse(reader *bufio.Reader) (*Response, error) {

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimRight(line, "\r\n")

	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf(
			"invalid HTTP response: %q",
			line,
		)
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

	var contentLength int64

	if v, ok := headers["Content-Length"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			contentLength = n
		}
	}

	chunked := false

	if v, ok := headers["Transfer-Encoding"]; ok {
		chunked = strings.Contains(
			strings.ToLower(v),
			"chunked",
		)
	}

	return &Response{
		Version:       parts[0],
		StatusCode:    code,
		Status:        parts[2],
		Headers:       headers,
		ContentLength: contentLength,
		Chunked:       chunked,
	}, nil
}