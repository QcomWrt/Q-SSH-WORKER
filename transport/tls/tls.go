package tls

import (
	cryptoTLS "crypto/tls"
	"fmt"
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

func Wrap(cfg *config.Config, conn net.Conn) (net.Conn, error) {
	if !cfg.Proxy.TLSEnable {
		return conn, nil
	}

	sniHost := cfg.Proxy.SNI
	if sniHost == "" {
		sniHost = cfg.Proxy.Host
	}

	tlsConfig := &cryptoTLS.Config{
		ServerName:         sniHost,
		InsecureSkipVerify: true,
	}

	tlsConn := cryptoTLS.Client(conn, tlsConfig)

	// Jalankan TLS Handshake secara eksplisit
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("tls handshake failed (SNI: %s): %w", sniHost, err)
	}

	return tlsConn, nil
}