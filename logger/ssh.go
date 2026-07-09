package logger

func SSHConnecting() {
    emit("Connecting SSH...")
}

func SSHConnected() {
    emit("SSH Connected")
}

func SSHError(err error) {
    emit("[SSH ERROR] Auth/Handshake failed")
}