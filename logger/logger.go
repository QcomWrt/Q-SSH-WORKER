package logger

import "fmt"

var Enable = true

func emit(msg string) {

	if !Enable {
		return
	}

	fmt.Println(msg)
	writeFile(msg)
}

func Status(format string, args ...any) {
	emit(fmt.Sprintf(format, args...))
}
