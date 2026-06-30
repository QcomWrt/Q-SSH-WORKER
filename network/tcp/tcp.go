package direct

import (
	"context"
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/transport/dialer"
)

type Direct struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Direct {
	return &Direct{
		cfg: cfg,
	}
}

func (d *Direct) Dial(ctx context.Context) (net.Conn, error) {

	ips, err := dialer.Resolve(d.cfg.SSH.Host)
	if err != nil {
		return nil, err
	}

	return dialer.DialFirst(
		ctx,
		ips,
		d.cfg.SSH.Port,
	)
}