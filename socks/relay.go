package socks

import (
	"context"
	"io"
	"net"
	"sync"

	qdebug "github.com/QcomWrt/Q-SSH-WORKER/debug"
)

// Relay mempertahankan API lama untuk kompatibilitas OpenWrt.
func Relay(client net.Conn, remote net.Conn) {
	RelayContext(context.Background(), client, remote)
}

type closeWriter interface {
	CloseWrite() error
}

type copyResult struct {
	dir string
	n   int64
	err error
}

// RelayContext melakukan relay dua arah. Saat klien selesai mengirim,
// hanya arah tulis ke remote yang ditutup. Saat remote selesai (atau
// error), kedua sisi ditutup.
func RelayContext(ctx context.Context, client net.Conn, remote net.Conn) {
	var closeOnce sync.Once
	closeBoth := func() {
		closeOnce.Do(func() {
			_ = client.Close()
			_ = remote.Close()
		})
	}

	stopClose := context.AfterFunc(ctx, closeBoth)
	defer stopClose()
	defer closeBoth()

	done := make(chan copyResult, 2)

	go func() {
		n, err := io.Copy(remote, client)
		done <- copyResult{"client->remote", n, err}
	}()

	go func() {
		n, err := io.Copy(client, remote)
		done <- copyResult{"remote->client", n, err}
	}()

	var first copyResult
	select {
	case <-ctx.Done():
		qdebug.Printf("[SOCKS DEBUG] relay dibatalkan oleh context")
		return
	case first = <-done:
	}

	qdebug.Printf("[SOCKS DEBUG] relay selesai arah=%s bytes=%d err=%v",
		first.dir, first.n, first.err)

	// Klien selesai mengirim dan tanpa error: beri tahu remote lewat
	// half-close, lalu tunggu arah balasan selesai.
	if first.dir == "client->remote" && first.err == nil {
		if cw, ok := remote.(closeWriter); ok {
			_ = cw.CloseWrite()
		}

		select {
		case second := <-done:
			qdebug.Printf("[SOCKS DEBUG] relay selesai arah=%s bytes=%d err=%v",
				second.dir, second.n, second.err)
		case <-ctx.Done():
		}
		return
	}

	// Remote menutup (EOF) atau error: tutup keduanya.
}