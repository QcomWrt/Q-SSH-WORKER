package debug

import (
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

func SSHBanner(banner string) {
	Line("%s", strings.TrimRight(banner, "\r\n"))
}

func SSHServerMessage(msg string) {
	if !Enable {
		return
	}
	Line("Server Message:")
	for _, line := range strings.Split(strings.TrimRight(msg, "\r\n"), "\n") {
		Line("%s", strings.TrimRight(line, "\r"))
	}
}

func EvaluateRejectResponse(statusCode int, status string, headers map[string]string, conn net.Conn, reader io.Reader) {
	if !Enable {
		return
	}
	Line("HTTP %d %s", statusCode, status)

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	var body strings.Builder
	if cl, ok := headers["Content-Length"]; ok {
		if n, err := strconv.Atoi(cl); err == nil && n > 0 {
			buf := make([]byte, n)
			_, _ = io.ReadFull(reader, buf)
			body.Write(buf)
		}
	} else {
		buf := make([]byte, 1024)
		if n, err := reader.Read(buf); err == nil && n > 0 {
			body.Write(buf[:n])
		}
	}

	if msg := body.String(); msg != "" {
		SSHServerMessage(msg)
	}
}