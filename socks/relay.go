package socks

import (
	"io"
	"net"
)

// Relay melakukan penyalinan data timbal balik (Piping) secara bidirectional
func Relay(client net.Conn, remote net.Conn) {
	errChan := make(chan error, 2)

	// Kirim: Aplikasi Lokal -> SSH Client -> VPS -> Internet
	go func() {
		_, err := io.Copy(remote, client)
		errChan <- err
	}()

	// Terima: Internet -> VPS -> SSH Client -> Aplikasi Lokal
	go func() {
		_, err := io.Copy(client, remote)
		errChan <- err
	}()

	// Blokir goroutine hingga salah satu jalur putus
	<-errChan
}