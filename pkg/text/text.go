package text

import (
	"regexp"
	"strings"
)

var (
	regexNewLine = regexp.MustCompile(`\n\s+`)
)

func Normalize(text string) string {
	text = regexNewLine.ReplaceAllString(text, "\a")
	text = strings.Join(strings.Fields(text), " ")
	text = strings.ReplaceAll(text, "\a", "\n")

	return text
}
