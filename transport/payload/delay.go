package payload

import "time"

func Delay(ms int) {

	if ms <= 0 {
		return
	}

	time.Sleep(time.Duration(ms) * time.Millisecond)
}