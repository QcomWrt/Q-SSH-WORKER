package payload

func Repeat(count int, fn func() error) error {

	if count <= 0 {
		count = 1
	}

	for i := 0; i < count; i++ {

		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}