package http

import (
	"bytes"
	"fmt"
	"sort"
)

// BuildRequest membangun request HTTP dari komponen-komponennya.
func BuildRequest(
	method string,
	target string,
	version string,
	headers map[string]string,
	body []byte,
) []byte {

	if version == "" {
		version = "HTTP/1.1"
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s %s %s\r\n", method, target, version)

	// Urutkan header agar output konsisten
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(&buf, "%s: %s\r\n", k, headers[k])
	}

	buf.WriteString("\r\n")

	if len(body) > 0 {
		buf.Write(body)
	}

	return buf.Bytes()
}