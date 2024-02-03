package contextualize

import (
	_ "embed"

	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
)

var (
	//go:embed prompt.tmpl
	promptTemplateText string
	promptTemplate     = prompt.MustNew(promptTemplateText)
)

type promptData struct {
	Input    string
	Messages []provider.Message
}
