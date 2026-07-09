package logger

import "fmt"

var Enable = true

func emit(msg string) {

	if !Enable {
		return
	}

	fmt.Println(msg)
}