package payload

import (
	"net"
	"strings"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
	"github.com/QcomWrt/Q-SSH-WORKER/logger"
	httpx "github.com/QcomWrt/Q-SSH-WORKER/transport/http"
	"github.com/QcomWrt/Q-SSH-WORKER/transport/response"
)

func Inject(cfg *config.Config, conn net.Conn) (net.Conn, error) {

	target := httpx.Target{
		Host: cfg.SSH.Host,
		Port: cfg.SSH.Port,
	}

	// 1. Render payload JSON menjadi string utuh mentah
	fullPayload := httpx.Render(
		cfg,
		target,
		cfg.Payload.Request,
	)

	var firstPart string
	var secondPart string

	// 2. Potong string payload secara presisi menggunakan standard strings Go
	if strings.Contains(fullPayload, "[split]") {
		parts := strings.SplitN(fullPayload, "[split]", 2)
		firstPart = parts[0]
		secondPart = parts[1]
	} else {
		// Fallback jika tidak ada tag [split], kirim semua di awal
		firstPart = fullPayload
	}

	logger.PayloadSending()

	debug.Println("\n[PART 1 - SENT TO PROXY]")
	debug.Println(firstPart)

	// 3. Kirim potongan pertama (GET / ...) murni ke jaringan
	if _, err := conn.Write([]byte(firstPart)); err != nil {
		return nil, err
	}

	logger.PayloadAccepted()

	// 4. Teruskan potongan kedua (PATCH / ... beserta HTTP/ 69) ke Once
	return response.Once(cfg, conn, secondPart)
}