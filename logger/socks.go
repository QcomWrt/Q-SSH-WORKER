package logger

import (
	"fmt"
)

func StatusConnected() {
	emit("[STATUS] Connected")
}

func SOCKS5Listening(addr string) {
	// Menggunakan format string lalu dilempar ke fungsi emit bawaan proyekmu
	emit(fmt.Sprintf("SOCKS5 Server listening on %s", addr))
}

func StatusError(msg string) {
	emit(fmt.Sprintf("[ERROR] %s", msg))
}