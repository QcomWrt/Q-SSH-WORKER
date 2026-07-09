package debug

import "sort"

func RawResponse(raw string) {
	if !Enable {
		return
	}

	Printf("[RAW RESPONSE] %q\n", raw)
}

func Response(
	version string,
	statusCode int,
	status string,
	headers map[string]string,
) {
	if !Enable {
		return
	}

	Separator()

	Println("========== RESPONSE ==========")

	Printf(
		"%s %d %s\n",
		version,
		statusCode,
		status,
	)

	keys := make([]string, 0, len(headers))

	for k := range headers {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		Printf(
			"%s: %s\n",
			k,
			headers[k],
		)
	}

	Println("==============================")
}