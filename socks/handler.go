package socks

import (
	"io"
	"net"
	"strconv"

	gossh "golang.org/x/crypto/ssh"
)

// HandleRequest memproses handshake SOCKS5 dan menghubungkannya ke terowongan SSH
func HandleRequest(clientConn net.Conn, sshClient *gossh.Client) {
	defer clientConn.Close()

	buf := make([]byte, 256)

	// ---- TAHAP 1: NEGOSIASI AUTENTIKASI ----
	if _, err := io.ReadFull(clientConn, buf[:2]); err != nil {
		return
	}
	if buf[0] != 0x05 { // SOCKS Versi 5
		return
	}

	numMethods := int(buf[1])
	if _, err := io.ReadFull(clientConn, buf[:numMethods]); err != nil {
		return
	}

	// Tanggapi: NO AUTHENTICATION REQUIRED (0x00)
	if _, err := clientConn.Write([]byte{0x05, 0x00}); err != nil {
		return
	}

	// ---- TAHAP 2: MEMBACA REQUEST PERINTAH ----
	if _, err := io.ReadFull(clientConn, buf[:4]); err != nil {
		return
	}

	cmd := buf[1]  // 0x01 = CONNECT
	atyp := buf[3] // Tipe Alamat (0x01 = IPv4, 0x03 = Domain)

	if cmd != 0x01 {
		// Tolak jika bukan perintah CONNECT (Unsupported Command)
		_, _ = clientConn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	var targetHost string

	switch atyp {
	case 0x01: // IPv4
		if _, err := io.ReadFull(clientConn, buf[:4]); err != nil {
			return
		}
		targetHost = net.IP(buf[:4]).String()
	case 0x03: // Domain Name
		if _, err := io.ReadFull(clientConn, buf[:1]); err != nil {
			return
		}
		domainLen := int(buf[0])
		if _, err := io.ReadFull(clientConn, buf[:domainLen]); err != nil {
			return
		}
		targetHost = string(buf[:domainLen])
	default:
		// Address type not supported
		_, _ = clientConn.Write([]byte{0x05, 0x08, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	// Membaca Port (2 Byte)
	if _, err := io.ReadFull(clientConn, buf[:2]); err != nil {
		return
	}
	targetPort := int(buf[0])<<8 | int(buf[1])
	targetAddr := net.JoinHostPort(targetHost, strconv.Itoa(targetPort))

	// ---- TAHAP 3: DIAL VIA SSH VIRTUAL PIPE ----
	sshConn, err := sshClient.Dial("tcp", targetAddr)
	if err != nil {
		// Kirim respon 0x03 (Network Unreachable) jika VPS gagal melakukan dial keluar
		_, _ = clientConn.Write([]byte{0x05, 0x03, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer sshConn.Close()

	// Kirim respon SUKSES (0x00) ke client lokal, tanda pipa siap dialiri data
	if _, err := clientConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}); err != nil {
		return
	}

	// ---- TAHAP 4: RELAY KAN TRAFIK DATA DUA ARAH ----
	Relay(clientConn, sshConn)
}