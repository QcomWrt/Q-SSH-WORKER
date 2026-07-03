package ssh

import (
	"fmt"

	gossh "golang.org/x/crypto/ssh"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

func KeyboardInteractive(cfg *config.Config) (gossh.AuthMethod, error) {

	return nil, fmt.Errorf("keyboard-interactive authentication not implemented")
}