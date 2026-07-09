package dialer

import (
	"context"
	"net"
	"time"
)

func DialTCP(ctx context.Context, address string) (net.Conn, error) {
	return New(10 * time.Second).DialContext(ctx, "tcp", address)
}