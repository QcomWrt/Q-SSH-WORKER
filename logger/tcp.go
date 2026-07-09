package logger

func TCPConnecting() {
    emit("Connecting TCP...")
}

func TCPConnected() {
    emit("TCP Connected")
}