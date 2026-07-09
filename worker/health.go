package worker

import (
	"context"
	"time"

	"github.com/QcomWrt/Q-SSH-WORKER/logger"
	gossh "golang.org/x/crypto/ssh"
)

// MonitorHealth melakukan ping berkala ke IP/Domain publik melalui jaringan SSH Client.
// Jika koneksi mati gantung, ia akan mengirim sinyal true ke failChan.
func MonitorHealth(ctx context.Context, sshClient *gossh.Client, failChan chan<- bool) {
	// Gunakan interval 30 detik untuk pengecekan berkala
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	failureCount := 0
	maxFailures := 2 // Toleransi gagal 2 kali berturut-turut sebelum dianggap mati

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Lakukan dial internal lewat SSH ke endpoint ringan (misal DNS Google / Cloudflare)
			// Kita hanya cek apakah tcp handshake sukses menembus VPS atau tidak
			dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			
			// Dial dijalankan asinkronus agar tidak mengunci liveness check jika macet total
			done := make(chan bool, 1)
			go func() {
				conn, err := sshClient.Dial("tcp", "1.1.1.1:80")
				if err == nil {
					conn.Close()
					done <- true
				} else {
					done <- false
				}
			}()

			select {
			case <-dialCtx.Done():
				cancel()
				failureCount++
			case success := <-done:
				cancel()
				if success {
					failureCount = 0 // Reset jika sukses
				} else {
					failureCount++
				}
			}

			if failureCount >= maxFailures {
				logger.StatusError("Koneksi gantung dideteksi oleh Health Monitor!")
				failChan <- true
				return
			}
		}
	}
}