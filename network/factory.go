package network

import (
	"context"
	"fmt"
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/network/tcp"
	"github.com/QcomWrt/Q-SSH-WORKER/network/ws"
)

type Network interface {
	Dial(ctx context.Context) (net.Conn, error)
}

func New(cfg *config.Config) (Network, error) {

	switch cfg.Network.Type {

	case "tcp":
		return tcp.New(cfg), nil

	case "ws":
		return ws.New(cfg), nil

	default:
		return nil, fmt.Errorf(
			"unsupported network: %s",
			cfg.Network.Type,
		)
	}
}