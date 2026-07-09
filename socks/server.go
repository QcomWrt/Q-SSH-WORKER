package socks

import (
	"net"
	"strconv"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/logger"
	gossh "golang.org/x/crypto/ssh"
)

func ListenAndServe(cfg *config.Config, sshClient *gossh.Client) error {
	listenAddr := net.JoinHostPort(cfg.Listen.Host, strconv.Itoa(cfg.Listen.Port))

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	// Memicu emit: "SOCKS5 Server listening on 127.0.0.1:1080"
	logger.SOCKS5Listening(listenAddr)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			continue
		}

		go HandleRequest(clientConn, sshClient)
	}
}