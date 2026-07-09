package logger

import "fmt"

func Error(err error) {

    if err == nil {
        return
    }

    emit(fmt.Sprintf("Error: %v", err))
}