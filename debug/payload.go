package debug

import (
	"fmt"
	"strings"
)

// Payload mencetak payload dalam satu baris, dengan \r dan \n
// ditampilkan sebagai teks (seperti tampilan HTTP Custom).
func Payload(payload string) {
	if !Enable {
		return
	}
	line := strings.ReplaceAll(payload, "\r", `\r`)
	line = strings.ReplaceAll(line, "\n", `\n`)
	fmt.Println(line)
}