package ssh

import (
	gossh "golang.org/x/crypto/ssh"
)

func NewSession(client *gossh.Client) (*gossh.Session, error) {

	return client.NewSession()
}