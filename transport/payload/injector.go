package payload

import (
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

func Inject(cfg *config.Config, conn net.Conn) (net.Conn, error) {
	_ = cfg
	return conn, nil
}