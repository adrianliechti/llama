package text

import (
	"regexp"
	"strings"
)

func Normalize(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = regexp.MustCompile(`\n\s*\n\s*`).ReplaceAllString(text, "\a\a")
	text = regexp.MustCompile(`\n\s*`).ReplaceAllString(text, "\a")
	text = strings.Join(strings.Fields(text), " ")
	text = strings.ReplaceAll(text, "\a", "\n")

	return text
}
