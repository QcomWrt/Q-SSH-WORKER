package debug

import (
	"strings"
)

// SSHBanner mencetak banner versi biner dari server SSH target (e.g., SSH-2.0-dropbear)
func SSHBanner(banner string) {
	if !Enable {
		return
	}

	Separator()
	Println("========== SSH BANNER ==========")
	Println(strings.TrimRight(banner, "\r\n"))
	Println("================================")
}

// SSHServerMessage mencetak pesan selamat datang / MOTD tambahan dari VPS jika ada
func SSHServerMessage(msg string) {
	if !Enable {
		return
	}

	Separator()
	Println("========== SSH SERVER MESSAGE ==========")
	Printf("%s", msg) // Menggunakan Printf karena teks bawaan server sudah berformat
	Println("========================================")
}