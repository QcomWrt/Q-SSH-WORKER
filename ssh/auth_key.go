package ssh

import (
	"fmt"

	gossh "golang.org/x/crypto/ssh"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

func PrivateKey(cfg *config.Config) (gossh.AuthMethod, error) {

	return nil, fmt.Errorf("private key authentication not implemented")
}