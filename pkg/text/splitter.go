package text

import "strings"

func Split(text string) []string {
	parts := strings.Split(text, textSeperator(text))

	return parts
}

type Splitter struct {
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
