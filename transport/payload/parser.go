package payload

import "strings"

func Parse(payload string) []Action {

	parts := strings.Split(payload, "[split]")

	actions := make([]Action, 0, len(parts)*2)

	for _, part := range parts {

		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		actions = append(actions, Action{
			Type: ActionWrite,
			Data: part,
		})

		actions = append(actions, Action{
			Type: ActionRead,
		})
	}

	// Jika payload tidak mengandung [split],
	// tetap tunggu satu response.
	if len(actions) == 0 {

		payload = strings.TrimSpace(payload)

		actions = append(actions,
			Action{
				Type: ActionWrite,
				Data: payload,
			},
			Action{
				Type: ActionRead,
			},
		)
	}

	return actions
}