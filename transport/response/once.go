package response

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
	"github.com/QcomWrt/Q-SSH-WORKER/internal"
)

func Once(cfg *config.Config, conn net.Conn, secondPart string) (net.Conn, error) {
	reader := bufio.NewReader(conn)

	// 1. Baca baris status pertama (HTTP Status Line)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimRight(line, "\r\n")

	// Ekstrak komponen Status Line (Contoh: "HTTP/1.1 301 Moved Permanently")
	parts := strings.SplitN(line, " ", 3)
	version := "HTTP/1.1"
	statusCode := 0
	status := ""
	if len(parts) >= 2 {
		version = parts[0]
		statusCode, _ = strconv.Atoi(parts[1])
		if len(parts) == 3 {
			status = parts[2]
		}
	} else {
		status = line
	}

	// 2. Baca dan kumpulkan seluruh HTTP Headers ke dalam map
	headers := make(map[string]string)
	var headerLines []string
	for {
		hl, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		trimmedHl := strings.TrimRight(hl, "\r\n")
		if trimmedHl == "" {
			break // Batas akhir header didapatkan (\r\n\r\n)
		}
		headerLines = append(headerLines, hl) // Simpan untuk dikuras nanti jika diperlukan

		// Masukkan ke map key-value
		if idx := strings.Index(trimmedHl, ":"); idx != -1 {
			k := strings.TrimSpace(trimmedHl[:idx])
			v := strings.TrimSpace(trimmedHl[idx+1:])
			headers[k] = v
		}
	}

	// ======================================================================
	// 🟢 PANGGIL SUB-SYSTEM DEBUG RESPONSE ASLI MILIKMU
	// ======================================================================
	debug.Response(version, statusCode, status, headers)

	// SKENARIO A: RESPONS LANGSUNG SUKSES (101 / 200)
	if statusCode == 101 || statusCode == 200 {
		if debug.Enable {
			debug.Println("[SUCCESS] Terhubung ke gerbang WebSocket! Menguras header respons...")
			debug.Println("[SUCCESS] Buffer HTTP Bersih! Membaca Banner SSH...")
		}

		sshBanner, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		debug.SSHBanner(sshBanner)

		stream := io.MultiReader(strings.NewReader(sshBanner), reader, conn)
		return internal.NewBufferedConn(conn, stream), nil
	}

	// SKENARIO B: RESPONS PENGALIHAN MULTI-RESPONSE (301 / 302)
	if statusCode == 301 || statusCode == 302 {
		if debug.Enable {
			debug.Println("[TUNNEL TRICK] Mendeteksi status 301, menguras header pengalihan...")
		}

		if secondPart != "" {
			if debug.Enable {
				debug.Println("[TUNNEL TRICK] Mengirimkan sisa payload hasil [split]...")
			}
			if _, err := conn.Write([]byte(secondPart)); err != nil {
				return nil, err
			}
		}

		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if debug.Enable {
			debug.Println("[SWEEPING] Menyapu sisa bodi kotor untuk mencari gerbang biner SSH...")
		}

		for {
			peekLine, err := reader.ReadString('\n')
			if err != nil {
				_ = conn.SetReadDeadline(time.Time{})
				if reader.Buffered() > 0 {
					stream := io.MultiReader(reader, conn)
					return internal.NewBufferedConn(conn, stream), nil
				}
				return nil, err
			}

			if strings.HasPrefix(peekLine, "SSH-") {
				_ = conn.SetReadDeadline(time.Time{})
				
				debug.SSHBanner(peekLine)

				stream := io.MultiReader(strings.NewReader(peekLine), reader, conn)
				return internal.NewBufferedConn(conn, stream), nil
			}
		}
	}

	if reader.Buffered() > 0 {
		stream := io.MultiReader(reader, conn)
		return internal.NewBufferedConn(conn, stream), nil
	}
	return conn, nil
}