package logger

import (
	"fmt"
	"os"
	"sync"
)

var (
	fileMu  sync.Mutex
	logFile *os.File
)

// OpenFile membuka file log. Setelah dipanggil, emit juga menulis ke file.
// Dipanggil sekali saat start, misalnya dari main.go.
func OpenFile(path string) error {
	fileMu.Lock()
	defer fileMu.Unlock()

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	logFile = f
	return nil
}

// writeFile menulis satu baris ke file log, jika file sudah dibuka.
func writeFile(msg string) {
	fileMu.Lock()
	defer fileMu.Unlock()

	if logFile == nil {
		return
	}
	fmt.Fprintln(logFile, msg)
}
