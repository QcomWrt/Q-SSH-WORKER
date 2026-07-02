package payload

import (
	"fmt"
	"net"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	httpx "github.com/QcomWrt/Q-SSH-WORKER/transport/http"
)

func Inject(cfg *config.Config, conn net.Conn) (net.Conn, error) {

	// Render template ([host], [port], [crlf], dst)
	payload := httpx.Render(
		cfg,
		cfg.Payload.Request,
	)

	fmt.Println("[PAYLOAD] Render")

	// Pecah berdasarkan [split]
	parts := Split(payload)

	fmt.Println("[PAYLOAD] Split")

	fmt.Printf("[PAYLOAD] Parts : %d\n", len(parts))

	// Kirim setiap bagian
	for i, part := range parts {

		fmt.Printf("[PAYLOAD] Write part %d\n", i+1)

		err := Write(conn, part)
		if err != nil {
			return nil, err
		}
	}

	// Baca response pertama
	resp, err := httpx.ReadResponse(conn)
	if err != nil {
		return nil, err
	}

	fmt.Println("[PAYLOAD] Read Response")

	fmt.Printf(
		"[PAYLOAD] Response : %d %s\n",
		resp.StatusCode,
		resp.Status,
	)

	fmt.Println("[PAYLOAD] Expect")
	// Validasi response
	if !expect(cfg, resp) {
		return nil, fmt.Errorf(
			"unexpected response: %d %s",
			resp.StatusCode,
			resp.Status,
		)
	}

	return conn, nil
}