package debug

func Part(index int, data string) {

	if !Enable {
		return
	}

	Printf(
		"\n----- PART %d -----\n",
		index,
	)

	Println(data)
}

func Bytes(n int) {

	if !Enable {
		return
	}

	Printf("[WRITE] %d bytes\n", n)
}