package proxy

import (
	"fmt"
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	httpx "github.com/QcomWrt/Q-SSH-WORKER/transport/http"
)

func Connect(cfg *config.Config, conn net.Conn) (net.Conn, error) {

	target := fmt.Sprintf("%s:%d",
		cfg.SSH.Host,
		cfg.SSH.Port,
	)

	req := httpx.BuildRequest(
		"CONNECT",
		target,
		"",
		map[string]string{
			"Host": target,
		},
		nil,
	)

	if _, err := conn.Write(req); err != nil {
		return nil, err
	}

	resp, err := httpx.ReadResponse(conn)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"proxy rejected: %d %s",
			resp.StatusCode,
			resp.Status,
		)
	}

	return conn, nil
}