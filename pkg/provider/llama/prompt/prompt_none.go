package prompt

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type promptNone struct {
}

func (t *promptNone) Prompt(system string, messages []provider.Message) (string, error) {
	var prompt string

	for _, message := range messages {
		if message.Role == provider.MessageRoleUser {
			prompt = strings.TrimSpace(message.Content)
		}
	}

	return prompt, nil
}

func (t *promptNone) Stop() []string {
	return []string{}
}
