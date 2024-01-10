package prompt

import (
	"errors"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Template interface {
	Prompt(system string, messages []provider.Message, options *TemplateOptions) (string, error)
	Stop() []string
}

type TemplateOptions struct {
	Functions []provider.Function
}

var (
	ErrFunctionsUnsupported = errors.New("functions are not supported")
)

var (
	None   Template = &promptNone{}
	Simple Template = &promptSimple{}

	ChatML Template = &promptChatML{}
	ToRA   Template = &promptToRA{}

	Llama      Template = &promptLlama{}
	LlamaGuard Template = &promptLlamaGuard{}
	Mistral    Template = &promptMistral{}
	NexusRaven Template = &promptNexusRaven{}
)
