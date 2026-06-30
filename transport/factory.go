package transport

import (
	"fmt"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/transport/dialer"
	"github.com/QcomWrt/Q-SSH-WORKER/transport/direct"
	// "github.com/QcomWrt/Q-SSH-WORKER/transport/http"
	// "github.com/QcomWrt/Q-SSH-WORKER/transport/tls"
	// "github.com/QcomWrt/Q-SSH-WORKER/transport/ws"
)

func New(cfg *config.Config) (dialer.Dialer, error) {

	switch cfg.Transport.Type {

	case "direct":
		return direct.New(cfg), nil

	// case "tls":
	//     return tls.New(cfg), nil

	// case "ws":
	//     return ws.New(cfg), nil

	default:
		return nil, fmt.Errorf("unsupported transport: %s", cfg.Transport.Type)
	}
}