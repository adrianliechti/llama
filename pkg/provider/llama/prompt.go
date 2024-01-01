package llama

import (
	"github.com/adrianliechti/llama/pkg/provider"
)

type PromptTemplate interface {
	Stop() []string
	Prompt(system string, messages []provider.Message) (string, error)
}
