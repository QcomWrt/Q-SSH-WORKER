package logger

import "fmt"

func SSHConnecting() {
	emit("Connecting SSH...")
}

func SSHConnected() {
	emit("SSH Connected")
}

func SSHError(err error) {
	if err == nil {
		emit("[SSH ERROR] Unknown error (nil)")
		return
	}

	emit(fmt.Sprintf("[SSH ERROR] %v", err))
}
