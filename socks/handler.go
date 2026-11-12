
package socks

import (
	"context"
	"io"
	"net"
	"strconv"

	gossh "golang.org/x/crypto/ssh"
	qdebug "github.com/QcomWrt/Q-SSH-WORKER/debug"
)

// HandleRequest mempertahankan API lama untuk kompatibilitas OpenWrt.
func HandleRequest(clientConn net.Conn, sshClient *gossh.Client) {
	HandleRequestContext(context.Background(), clientConn, sshClient)
}

// HandleRequestContext memproses handshake SOCKS5 dan menghubungkan
// koneksi klien ke tujuan melalui SSH.
func HandleRequestContext(
	ctx context.Context,
	clientConn net.Conn,
	sshClient *gossh.Client,
) {
	defer clientConn.Close()

	// Hentikan operasi baca/tulis klien ketika context dibatalkan.
	stopClose := context.AfterFunc(ctx, func() {
		_ = clientConn.Close()
	})
	defer stopClose()

	if ctx.Err() != nil {
		return
	}

	buf := make([]byte, 256)

	// --------------------------------------------------
	// TAHAP 1: NEGOSIASI AUTENTIKASI SOCKS5
	// --------------------------------------------------
	if _, err := io.ReadFull(clientConn, buf[:2]); err != nil {
		return
	}

	if buf[0] != 0x05 {
		return
	}

	numMethods := int(buf[1])
	if numMethods == 0 {
		_, _ = clientConn.Write([]byte{0x05, 0xFF})
		return
	}

	if _, err := io.ReadFull(clientConn, buf[:numMethods]); err != nil {
		return
	}

	// Periksa apakah klien mendukung metode tanpa autentikasi.
	noAuthSupported := false
	for _, method := range buf[:numMethods] {
		if method == 0x00 {
			noAuthSupported = true
			break
		}
	}

	if !noAuthSupported {
		_, _ = clientConn.Write([]byte{0x05, 0xFF})
		return
	}

	if _, err := clientConn.Write([]byte{0x05, 0x00}); err != nil {
		return
	}

	if ctx.Err() != nil {
		return
	}

	// --------------------------------------------------
	// TAHAP 2: MEMBACA REQUEST CONNECT
	// --------------------------------------------------
	if _, err := io.ReadFull(clientConn, buf[:4]); err != nil {
		return
	}

	if buf[0] != 0x05 {
		return
	}

	cmd := buf[1]
	atyp := buf[3]

	if cmd != 0x01 {
		// Command tidak didukung.
		_, _ = clientConn.Write(
			[]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0},
		)
		return
	}

	var targetHost string

	switch atyp {
	case 0x01: // IPv4
		if _, err := io.ReadFull(clientConn, buf[:4]); err != nil {
			return
		}
		targetHost = net.IP(buf[:4]).String()

	case 0x03: // Domain name
		if _, err := io.ReadFull(clientConn, buf[:1]); err != nil {
			return
		}

		domainLen := int(buf[0])
		if domainLen == 0 {
			return
		}

		if _, err := io.ReadFull(clientConn, buf[:domainLen]); err != nil {
			return
		}
		targetHost = string(buf[:domainLen])

	case 0x04: // IPv6
		if _, err := io.ReadFull(clientConn, buf[:16]); err != nil {
			return
		}
		targetHost = net.IP(buf[:16]).String()

	default:
		// Address type tidak didukung.
		_, _ = clientConn.Write(
			[]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0},
		)
		return
	}

	// Port tujuan terdiri dari dua byte big-endian.
	if _, err := io.ReadFull(clientConn, buf[:2]); err != nil {
		return
	}

	targetPort := int(buf[0])<<8 | int(buf[1])
	targetAddr := net.JoinHostPort(
		targetHost,
		strconv.Itoa(targetPort),
	)

	if ctx.Err() != nil {
		return
	}

	// --------------------------------------------------
	// TAHAP 3: DIAL MELALUI SSH
	// --------------------------------------------------
	sshConn, err := sshClient.Dial("tcp", targetAddr)
	if err != nil {
		qdebug.Printf("[SOCKS DEBUG] ssh dial %s gagal: %v", targetAddr, err)
		if ctx.Err() != nil {
			return
		}

		_, _ = clientConn.Write(
			[]byte{0x05, 0x03, 0x00, 0x01, 0, 0, 0, 0, 0, 0},
		)
		return
	}
	defer sshConn.Close()

	if ctx.Err() != nil {
		return
	}

	// --------------------------------------------------
	// TAHAP 4: KIRIM RESPON SUKSES
	// --------------------------------------------------
	if _, err := clientConn.Write(
		[]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0},
	); err != nil {
		return
	}

	// --------------------------------------------------
	// TAHAP 5: RELAY DATA DUA ARAH
	// --------------------------------------------------

	RelayContext(ctx, clientConn, sshConn)
}
