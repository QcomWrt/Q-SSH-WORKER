package ssh

import (
	"errors"
	"fmt"
	"net"
	"time"

	gossh "golang.org/x/crypto/ssh"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
)

const (
	// keepaliveInterval: jeda antar keepalive SSH ke server.
	keepaliveInterval = 5 * time.Second
	// keepaliveTimeout: batas menunggu balasan keepalive.
	// Sesi dianggap mati kalau tidak ada balasan dalam batas ini.
	keepaliveTimeout = 5 * time.Second
)

// errKeepaliveTimeout dipakai saat server tidak membalas keepalive.
var errKeepaliveTimeout = errors.New("keepalive timeout: server tidak membalas")

func Dial(cfg *config.Config, conn net.Conn) (*gossh.Client, error) {

	authMethods := []gossh.AuthMethod{}

	// Password
	if auth, err := Password(cfg); err == nil {
		authMethods = append(authMethods, auth)
	}

	// Private Key
	if auth, err := PrivateKey(cfg); err == nil {
		authMethods = append(authMethods, auth)
	}

	// Keyboard Interactive
	if auth, err := KeyboardInteractive(cfg); err == nil {
		authMethods = append(authMethods, auth)
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no ssh authentication method available")
	}

	clientConfig := &gossh.ClientConfig{
		User:            cfg.SSH.Username,
		Auth:            authMethods,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),

		// Menangkap pesan banner/MOTD dari VPS
		BannerCallback: func(message string) error {
			debug.SSHServerMessage(message)
			return nil
		},
	}

	cc, chans, reqs, err := gossh.NewClientConn(
		conn,
		conn.RemoteAddr().String(),
		clientConfig,
	)
	if err != nil {
		return nil, err
	}

	client := gossh.NewClient(cc, chans, reqs)

	// Keepalive: deteksi sesi mati lebih cepat daripada probe SOCKS.
	go keepAlive(client)

	return client, nil
}

// keepAlive mengirim request keepalive berkala. Jika server tidak membalas
// dalam keepaliveTimeout, client ditutup. Penutupan itu membuat client.Wait()
// kembali, sehingga sesi dinyatakan putus.
func keepAlive(client *gossh.Client) {
	ticker := time.NewTicker(keepaliveInterval)
	defer ticker.Stop()

	for range ticker.C {
		result := make(chan error, 1)
		go func() {
			_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
			result <- err
		}()

		select {
		case err := <-result:
			if err != nil {
				// Koneksi sudah tertutup dari luar, atau rusak.
				return
			}
		case <-time.After(keepaliveTimeout):
			_ = client.Close()
			_ = errKeepaliveTimeout
			return
		}
	}
}
