package logger

import (
	"fmt"
)

func ProxyConnecting() {
	emit("Connecting Proxy...")
}

func ProxyConnected() {
	emit("Proxy Connected")
}

func ProxyError(err error) {
	emit(fmt.Sprintf("[PROXY ERROR] %v", err))
}
