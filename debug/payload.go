package debug

func Payload(payload string) {

	if !Enable {
		return
	}

	Separator()

	Println("========== PAYLOAD ==========")
	Println(payload)
	Println("=============================")
}