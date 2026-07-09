package debug

func ExpectMatched(kind string, value string) {

	if !Enable {
		return
	}

	Printf(
		"[EXPECT] matched %s : %s\n",
		kind,
		value,
	)
}

func ExpectRejected(status string) {

	if !Enable {
		return
	}

	Printf(
		"[EXPECT] rejected : %s\n",
		status,
	)
}

func ExpectSkipped() {

	if !Enable {
		return
	}

	Println("[EXPECT] skipped")
}