package proxy

import (
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

func Connect(cfg *config.Config, conn net.Conn) (net.Conn, error) {
	_ = cfg
	return conn, nil
}