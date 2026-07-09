package dialer

import (
	"strconv"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

func Address(cfg *config.Config) string {

	if cfg.Proxy.Enable &&
		cfg.Proxy.Host != "" &&
		cfg.Proxy.Port > 0 {

		return cfg.Proxy.Host + ":" +
			strconv.Itoa(cfg.Proxy.Port)
	}

	return cfg.SSH.Host + ":" +
		strconv.Itoa(cfg.SSH.Port)
}