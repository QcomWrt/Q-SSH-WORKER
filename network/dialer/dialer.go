package dialer

import (
	"net"
	"syscall"
	"time"
)

var Control func(network, address string, c syscall.RawConn) error

func New(timeout time.Duration) *net.Dialer {
	return &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
		Control:   Control,
	}
}