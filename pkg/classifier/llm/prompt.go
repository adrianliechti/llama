package llm

import (
	_ "embed"

	"github.com/adrianliechti/llama/pkg/prompt"
)

var (
	//go:embed prompt.tmpl
	promptTemplateText string
	promptTemplate     = prompt.MustNew(promptTemplateText)

	promptStop = []string{
		"\n###",
		"\nClass:",
	}
)

type promptData struct {
	Input      string
	Categories []Category
}
