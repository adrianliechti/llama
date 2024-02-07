package summarize

import (
	_ "embed"
)

var (
	//go:embed prompt.tmpl
	promptTemplate string
)

type promptData struct {
	Input string
}
