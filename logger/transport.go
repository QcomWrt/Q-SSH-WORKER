package logger

func TransportConnecting() {
    emit("Connecting Transport...")
}

func TransportConnected() {
    emit("Transport Connected")
}