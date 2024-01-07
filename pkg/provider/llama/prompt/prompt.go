package prompt

import (
	"github.com/adrianliechti/llama/pkg/provider"
)

type Template interface {
	Prompt(system string, messages []provider.Message) (string, error)
	Stop() []string
}

var (
	Simple     Template = &promptSimple{}
	ChatML     Template = &promptChatML{}
	Llama      Template = &promptLlama{}
	LlamaGuard Template = &promptLlamaGuard{}
	Mistral    Template = &promptMistral{}
)
