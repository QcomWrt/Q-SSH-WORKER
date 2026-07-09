package debug

func Split(raw, valid int) {

	if !Enable {
		return
	}

	Printf("[SPLIT] Raw parts : %d\n", raw)
	Printf("[SPLIT] Valid parts : %d\n", valid)
}