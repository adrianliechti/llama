package text

import (
	"strings"
)

const chunkSize = 1500

func Split(text string) []string {
	text = Normalize(text)

	sep := textSeperator(text)
	parts := strings.Split(text, sep)

	var chunks []string
	var current []string

	for _, part := range parts {
		if len(part) < chunkSize {
			current = append(current, part)
			continue
		}

		if len(current) > 0 {
			chunks = append(chunks, combineChunks(current, sep)...)
			clear(current)
		}

		chunks = append(chunks, Split(part)...)
	}

	if len(current) > 0 {
		chunks = append(chunks, combineChunks(current, sep)...)
	}

	var result []string

	for _, c := range chunks {
		c := strings.TrimSpace(c)

		if c == "" {
			continue
		}

		result = append(result, c)
	}

	return result
}

func combineChunks(chunks []string, sep string) []string {
	var result []string

	var chunk string

	for _, c := range chunks {
		if len(chunk)+len(c) > chunkSize {
			result = append(result, chunk)
			chunk = ""
		}

		chunk += sep + c
	}

	if chunk != "" {
		result = append(result, chunk)
	}

	return result
}

func textSeperator(text string) string {
	separators := []string{
		"\n\n",
		"\n",
		" ",
		"",
	}

	for _, sep := range separators {
		if strings.Contains(text, sep) {
			return sep
		}
	}

	return ""
}
