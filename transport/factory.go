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

	// 1. Proxy Layer Execution
	if cfg.Proxy.Enable {
		conn, err = proxy.Connect(cfg, conn)
		if err != nil {
			return nil, err
		}
	}

	// 2. TLS Layer Execution (TLS Handshake dilakukan SEBELUM Payload)
	if cfg.Transport.TLS {
		conn, err = tls.Wrap(cfg, conn)
		if err != nil {
			return nil, err
		}
	}

	// 3. Payload Injection Layer Execution
	if cfg.Payload.Enable {
		conn, err = payload.Inject(cfg, conn)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}