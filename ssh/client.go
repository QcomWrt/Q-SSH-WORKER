package ssh

import (
	"fmt"
	"net"

	gossh "golang.org/x/crypto/ssh"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug" // Import modul debug milikmu
)

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

	// Inisialisasi clientConfig menggunakan alias gossh yang benar
	clientConfig := &gossh.ClientConfig{
		User: cfg.SSH.Username,
		Auth: authMethods, // Menggunakan penanganan authMethods yang dinamis bawaanmu

		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		
		// KUNCI UTAMA: Menangkap pesan banner/MOTD dari VPS secara legal
		BannerCallback: func(message string) error {
			debug.SSHServerMessage(message)
			return nil
		},
	}

	cc, chans, reqs, err := gossh.NewClientConn(
		conn,
		conn.RemoteAddr().String(),
		clientConfig, // Sekarang variabel ini sudah terdefinisi secara presisi
	)
	if err != nil {
		return nil, err
	}

	return gossh.NewClient(
		cc,
		chans,
		reqs,
	), nil
}