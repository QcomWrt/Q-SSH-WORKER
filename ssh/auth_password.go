package ssh

import (
	"fmt"

	gossh "golang.org/x/crypto/ssh"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

func Password(cfg *config.Config) (gossh.AuthMethod, error) {

	if cfg.SSH.Username == "" {
		return nil, fmt.Errorf("ssh username is empty")
	}

	if cfg.SSH.Password == "" {
		return nil, fmt.Errorf("ssh password is empty")
	}

	return gossh.Password(cfg.SSH.Password), nil
}