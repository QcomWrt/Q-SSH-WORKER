package debug

import "fmt"

var Enable = true

func Println(a ...any) {
	if Enable {
		fmt.Println(a...)
	}
}

func Printf(format string, a ...any) {
	if Enable {
		fmt.Printf(format, a...)
	}
}

func Separator() {
	if Enable {
		fmt.Println()
	}
}