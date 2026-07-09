package http

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func Discard(reader *bufio.Reader, resp *Response) error {

	if strings.EqualFold(resp.Headers["Transfer-Encoding"], "chunked") {

		for {

			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}

			line = strings.TrimSpace(line)

			size, err := strconv.ParseInt(line, 16, 64)
			if err != nil {
				return fmt.Errorf("invalid chunk size: %q", line)
			}

			if size == 0 {

				_, err = reader.ReadString('\n')
				return err
			}

			if _, err := io.CopyN(io.Discard, reader, size); err != nil {
				return err
			}

			if _, err := reader.ReadString('\n'); err != nil {
				return err
			}
		}
	}

	if v := resp.Headers["Content-Length"]; v != "" {

		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}

		_, err = io.CopyN(io.Discard, reader, n)

		return err
	}

	return nil
}