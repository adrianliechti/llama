package template

import (
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/template/mistral"
)

type Template interface {
	Stop() []string

	Render(messages []provider.Message, options *provider.CompleteOptions) (string, error)
}

var (
	Mistral Template = &mistral.Template{}
)
