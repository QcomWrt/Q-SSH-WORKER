package response

import (
	"strconv"
	"strings"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
	"github.com/QcomWrt/Q-SSH-WORKER/debug"
	httpx "github.com/QcomWrt/Q-SSH-WORKER/transport/http"
)

func Expect(cfg *config.Config, resp *httpx.Response) bool {

	if len(cfg.Payload.Expect) == 0 {
		debug.ExpectSkipped()
		return true
	}

	status := strconv.Itoa(resp.StatusCode)

	for _, e := range cfg.Payload.Expect {

		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}

		// Match status code
		if status == e {
			debug.ExpectMatched("status code", e)
			return true
		}

		// Match status text
		if strings.Contains(
			strings.ToLower(resp.Status),
			strings.ToLower(e),
		) {
			debug.ExpectMatched("status text", e)
			return true
		}
	}

	debug.ExpectRejected(resp.Status)

	return false
}