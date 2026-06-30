package dialer

import (
	"context"
	"net"
)

type Dialer interface {
	Dial(ctx context.Context) (net.Conn, error)
}