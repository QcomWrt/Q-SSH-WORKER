package internal

import (
	"fmt"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

func Address(cfg *config.Config) string {
	return fmt.Sprintf("%s:%d", cfg.SSH.Host, cfg.SSH.Port)
}