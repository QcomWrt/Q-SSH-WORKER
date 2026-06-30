package dialer

import (
	"context"
	"fmt"
	"net"
	"strconv"
)

func DialTCP(ctx context.Context, host string, port int) (net.Conn, error) {

	addr := net.JoinHostPort(host, strconv.Itoa(port))

	var d net.Dialer

	return d.DialContext(ctx, "tcp", addr)
}

func DialFirst(ctx context.Context, ips []string, port int) (net.Conn, error) {

	var last error

	for _, ip := range ips {

		conn, err := DialTCP(ctx, ip, port)
		if err == nil {
			return conn, nil
		}

		last = err
	}

	if last != nil {
		return nil, last
	}

	return nil, fmt.Errorf("no endpoint available")
}