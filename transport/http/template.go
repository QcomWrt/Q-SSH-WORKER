package http

import (
	"strconv"
	"strings"

	"github.com/QcomWrt/Q-SSH-WORKER/config"
)

const (
	CRLF = "\r\n"
	LF   = "\n"
	CR   = "\r"
)

// Render mengganti placeholder pada template HTTP.
func Render(cfg *config.Config, target Target, text string) string {

	replacer := strings.NewReplacer(

		// Line ending
		"[crlf]", CRLF,
		"[lf]", LF,
		"[cr]", CR,

		// SSH target
		"[host]", target.Host,
		"[port]", strconv.Itoa(target.Port),
		"[host_port]", target.Host+":"+strconv.Itoa(target.Port),

		// HTTP transport
		"[path]", cfg.Transport.Path,

		// Protocol
		"[protocol]", protocol(cfg),

		// User-Agent
		"[ua]", defaultUserAgent(),
	)

	return replacer.Replace(text)
}

func protocol(cfg *config.Config) string {

	if cfg.Transport.TLS {
		return "https"
	}

	return "http"
}

func defaultUserAgent() string {

	return "Q-SSH-WORKER/1.0"
}