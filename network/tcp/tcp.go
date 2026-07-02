package tcp

import (
	"context"
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/network/dialer"
)

type TCP struct {
	cfg *config.Config
}

func New(cfg *config.Config) *TCP {
	return &TCP{
		cfg: cfg,
	}
}

func (t *TCP) Dial(ctx context.Context) (net.Conn, error) {

	return dialer.DialTCP(
		ctx,
		dialer.Address(t.cfg),
	)
}