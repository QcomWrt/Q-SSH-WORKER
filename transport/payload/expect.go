package payload

import (
	"strconv"
	"strings"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	httpx "github.com/QcomWrt/Q-SSH-WORKER/transport/http"
)

func expect(cfg *config.Config, resp *httpx.Response) bool {

	if len(cfg.Payload.Expect) == 0 {
		return true
	}

	status := strconv.Itoa(resp.StatusCode)

	for _, e := range cfg.Payload.Expect {

		e = strings.TrimSpace(e)

		if e == "" {
			continue
		}

		if status == e {
			return true
		}

		if strings.Contains(
			strings.ToLower(resp.Status),
			strings.ToLower(e),
		) {
			return true
		}
	}

	return false
}