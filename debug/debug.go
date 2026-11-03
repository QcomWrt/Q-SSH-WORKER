package debug

import (
	"fmt"
	"strings"
)

var Enable = false

func Println(a ...any) {
	if Enable {
		fmt.Println(a...)
	}
}

// Printf mencetak detail debug. Newline ditambahkan otomatis bila format
// belum diakhiri baris baru, sehingga pemanggil tidak perlu mengingatnya.
func Printf(format string, a ...any) {
	if !Enable {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Printf(format, a...)
}

func Separator() {
	if Enable {
		fmt.Println()
	}
}
// Line mencetak satu baris debug dengan prefix [DEBUG].
// Line mencetak satu baris debug. Newline ditambahkan otomatis.
func Line(format string, a ...any) {
	if !Enable {
		return
	}
	fmt.Printf(format+"\n", a...)
}