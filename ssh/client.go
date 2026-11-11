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
	// Default nilai parameter Keep-Alive untuk SSH mode biasa (TCP Direct/Payload/SNI)
	keepaliveInterval = 15 * time.Second
	keepaliveTimeout  = 30 * time.Second
	maxKeepaliveFails = 3
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

	// PISAHKAN LOGIKA KEEPALIVE:
	// Mode SSH Biasa (TCP Direct/Payload/SNI) -> Jalankan keepAlive agar socket idle tidak diputus ISP.
	// Mode SlowDNS (cfg.DNS.Enabled() == true) -> Nonaktifkan SSH Keep-Alive untuk cegah false disconnect saat stream data besar.
	if !cfg.DNS.Enabled() {
		go keepAlive(client)
	} else {
		debug.Printf("[SSH DEBUG] SSH Keep-Alive dinonaktifkan khusus untuk transport SlowDNS (TargetDomain: %s)", cfg.DNS.TargetDomain)
	}

	return client, nil
}

// keepAlive mengirim request keepalive berkala secara tertata tanpa leaking goroutine.
func keepAlive(client *gossh.Client) {
	ticker := time.NewTicker(keepaliveInterval)
	defer ticker.Stop()

	fails := 0

	for range ticker.C {
		type pingResult struct {
			err error
		}

		ch := make(chan pingResult, 1)
		go func() {
			_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
			ch <- pingResult{err: err}
		}()

		select {
		case res := <-ch:
			if res.err != nil {
				fails++
				debug.Printf("[SSH DEBUG] Keepalive ping error (%d/%d): %v", fails, maxKeepaliveFails, res.err)
			} else {
				fails = 0 // Reset hitungan jika berhasil menerima balasan
			}
		case <-time.After(keepaliveTimeout):
			fails++
			debug.Printf("[SSH DEBUG] Keepalive ping timeout (%d/%d)", fails, maxKeepaliveFails)
		}

		if fails >= maxKeepaliveFails {
			debug.Printf("[SSH DEBUG] Memutus koneksi SSH karena %d kali keepalive gagal berturut-turut", fails)
			_ = client.Close()
			return
		}
	}
}