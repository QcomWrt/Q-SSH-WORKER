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

	// Proxy Layer Execution
	if cfg.Proxy.Enable {
		// ❌ logger.ProxyConnecting() DIHAPUS DARI SINI
		conn, err = proxy.Connect(cfg, conn)
		if err != nil {
			return nil, err
		}
		// ❌ logger.ProxyConnected() DIHAPUS DARI SINI
	}

	// Payload Injection Layer Execution
	if cfg.Payload.Enable {
		conn, err = payload.Inject(cfg, conn)
		if err != nil {
			return nil, err
		}
	}

	// TLS Layer Execution
	if cfg.Transport.TLS {
		conn, err = tls.Wrap(cfg, conn)
		if err != nil {
			return nil, err
		}
	}

	return conn, nil
}