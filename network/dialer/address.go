package dialer

import (
	"strconv"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

func Address(cfg *config.Config) string {

	if cfg.Payload.Enable &&
		cfg.Payload.ProxyHost != "" &&
		cfg.Payload.ProxyPort > 0 {

		return cfg.Payload.ProxyHost + ":" +
			strconv.Itoa(cfg.Payload.ProxyPort)
	}

	return cfg.SSH.Host + ":" +
		strconv.Itoa(cfg.SSH.Port)
}