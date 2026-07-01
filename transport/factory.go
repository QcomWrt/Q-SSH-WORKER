package transport

import (
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"

	"github.com/QcomWrt/Q-SSH-WORKER/transport/payload"
	"github.com/QcomWrt/Q-SSH-WORKER/transport/proxy"
	"github.com/QcomWrt/Q-SSH-WORKER/transport/tls"
)

func Wrap(cfg *config.Config, conn net.Conn) (net.Conn, error) {

	var err error

	// Proxy
	if cfg.Payload.Enable && cfg.Payload.ProxyHost != "" {
		conn, err = proxy.Connect(cfg, conn)
		if err != nil {
			return nil, err
		}
	}

	// Payload
	if cfg.Payload.Enable {
		conn, err = payload.Inject(cfg, conn)
		if err != nil {
			return nil, err
		}
	}

	// TLS
	if cfg.Transport.TLS {
		conn, err = tls.Wrap(cfg, conn)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}