package logger

import (
    "fmt"
)

func PayloadSending() {
    emit("Sending Payload...")
}

func PayloadAccepted() {
    emit("Payload Accepted")
}

func PayloadError(err error) {
    emit(fmt.Sprintf("[PAYLOAD ERROR] Injection failed: %v", err))
}