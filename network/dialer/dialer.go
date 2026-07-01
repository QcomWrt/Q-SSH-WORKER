package dialer

import (
	"net"
	"time"
)

func New(timeout time.Duration) *net.Dialer {

	return &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}
}