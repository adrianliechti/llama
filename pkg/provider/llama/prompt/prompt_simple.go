package prompt

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type promptSimple struct {
}

func (t *promptSimple) Prompt(system string, messages []provider.Message) (string, error) {
	if err := llamaMessageOrder(messages); err != nil {
		return "", err
	}

	if len(messages) > 0 && system != "" && messages[0].Role != provider.MessageRoleSystem {
		message := provider.Message{
			Role:    provider.MessageRoleSystem,
			Content: strings.TrimSpace(system),
		}

		messages = append([]provider.Message{message}, messages...)
	}

	var prompt strings.Builder

	for i, message := range messages {
		if message.Role == provider.MessageRoleSystem {
			if i == 0 {
				prompt.WriteString(strings.TrimSpace(message.Content))
				prompt.WriteString("\n")
			}
		}

		if message.Role == provider.MessageRoleUser {
			prompt.WriteString("USER: ")
			prompt.WriteString(strings.TrimSpace(message.Content))
			prompt.WriteString("\n")
		}

		if message.Role == provider.MessageRoleAssistant {
			prompt.WriteString("ASSSISTANT: ")
			prompt.WriteString(strings.TrimSpace(message.Content))
			prompt.WriteString("\n")

		}
	}

	prompt.WriteString("ASSSISTANT: ")

	return prompt.String(), nil
}

func (t *promptSimple) Stop() []string {
	return []string{
		"USER:",
		"ASSSISTANT:",
	}
}
