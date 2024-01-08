package refine

import (
	_ "embed"

	"github.com/adrianliechti/llama/pkg/prompt"
)

var (
	//go:embed prompt.tmpl
	promptTemplateText string
	promptTemplate     = prompt.MustNew(promptTemplateText)
)

type promptData struct {
	Input   string
	Context string

	Answer string
}
