package ws

import (
	"context"
	"errors"
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

type WebSocket struct {
	cfg *config.Config
}

func New(cfg *config.Config) *WebSocket {

	return &WebSocket{
		cfg: cfg,
	}
}

func (w *WebSocket) Dial(ctx context.Context) (net.Conn, error) {

	return nil, errors.New("websocket not implemented")
}