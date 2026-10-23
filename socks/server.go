
package socks

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/logger"
	gossh "golang.org/x/crypto/ssh"
)

// ListenAndServe mempertahankan API lama untuk OpenWrt.
func ListenAndServe(cfg *config.Config, sshClient *gossh.Client) error {
	return ListenAndServeContext(context.Background(), cfg, sshClient)
}

// ListenAndServeContext menjalankan SOCKS5 sampai context dibatalkan
// atau server mengalami error.
func ListenAndServeContext(
	ctx context.Context,
	cfg *config.Config,
	sshClient *gossh.Client,
) error {
	listenAddr := net.JoinHostPort(
		cfg.Listen.Host,
		strconv.Itoa(cfg.Listen.Port),
	)

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	// Tutup listener agar Accept() berhenti saat context dibatalkan.
	stopClose := context.AfterFunc(ctx, func() {
		_ = listener.Close()
	})
	defer stopClose()
	defer listener.Close()

	logger.SOCKS5Listening(listenAddr)

	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		clientConn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			// Error selain pembatalan harus diteruskan,
			// bukan diulang tanpa batas.
			if errors.Is(err, net.ErrClosed) {
				return err
			}
			return err
		}

		go func(conn net.Conn) {
			// HandleRequest bertanggung jawab atas koneksi klien.
			HandleRequest(conn, sshClient)
		}(clientConn)
	}
}
