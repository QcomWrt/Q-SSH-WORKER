
package socks

import (
	"context"
	"io"
	"net"
	"sync"
)

// Relay mempertahankan API lama untuk kompatibilitas OpenWrt.
func Relay(client net.Conn, remote net.Conn) {
	RelayContext(context.Background(), client, remote)
}

// RelayContext melakukan relay bidirectional dan menghentikan kedua
// arah transfer ketika context dibatalkan atau salah satu arah selesai.
func RelayContext(ctx context.Context, client net.Conn, remote net.Conn) {
	var closeOnce sync.Once

	closeBoth := func() {
		closeOnce.Do(func() {
			_ = client.Close()
			_ = remote.Close()
		})
	}

	// Membuka blokir io.Copy saat context dibatalkan.
	stopClose := context.AfterFunc(ctx, closeBoth)
	defer stopClose()
	defer closeBoth()

	errChan := make(chan error, 2)

	go func() {
		_, err := io.Copy(remote, client)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(client, remote)
		errChan <- err
	}()

	// Tunggu arah transfer pertama selesai.
	select {
	case <-ctx.Done():
	case <-errChan:
	}

	// Memutus kedua socket agar arah transfer lainnya tidak tertinggal.
	closeBoth()

	// Tunggu goroutine kedua menyelesaikan io.Copy.
	select {
	case <-errChan:
	case <-ctx.Done():
		// Socket sudah ditutup; context yang dibatalkan tidak perlu
		// membuat fungsi menunggu selamanya.
	}
}
