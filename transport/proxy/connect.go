package proxy

import (
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
)

func Connect(cfg *config.Config, conn net.Conn) (net.Conn, error) {

    debug.Proxy(
        cfg.Proxy.Host,
        cfg.Proxy.Port,
        cfg.SSH.Host,
        cfg.SSH.Port,
    )

	return conn, nil
}