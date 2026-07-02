package payload

import "strings"

const SplitToken = "[split]"

// Split memecah payload menjadi beberapa bagian
// berdasarkan token [split].
func Split(payload string) []string {

	parts := strings.Split(payload, SplitToken)

	out := make([]string, 0, len(parts))

	for _, part := range parts {

		part = strings.TrimSpace(part)

		if part == "" {
			continue
		}

		out = append(out, part)
	}

	return out
}