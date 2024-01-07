package prompt

import (
	"github.com/adrianliechti/llama/pkg/provider"
)

type Template interface {
	Prompt(system string, messages []provider.Message) (string, error)
	Stop() []string
}

var (
	None   Template = &promptNone{}
	Simple Template = &promptSimple{}

	ChatML Template = &promptChatML{}
	ToRA   Template = &promptToRA{}

	Llama      Template = &promptLlama{}
	LlamaGuard Template = &promptLlamaGuard{}
	Mistral    Template = &promptMistral{}
)
