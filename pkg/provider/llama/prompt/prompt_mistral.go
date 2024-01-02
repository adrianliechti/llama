package prompt

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type promptMistral struct {
}

func (t *promptMistral) Prompt(system string, messages []provider.Message) (string, error) {
	messages = llamaMessageFlattening(messages)

	if err := llamaMessageOrder(messages); err != nil {
		return "", err
	}

	if len(messages) > 0 && messages[0].Role == provider.MessageRoleSystem {
		system = strings.TrimSpace(messages[0].Content)
		messages = messages[1:]
	}

	var prompt strings.Builder

	for i, message := range messages {
		if message.Role == provider.MessageRoleUser {
			prompt.WriteString("[INST] ")

			if i == 0 && len(system) > 0 {
				prompt.WriteString(strings.TrimSpace(system))
				prompt.WriteString("\n\n")
			}

			if i > 0 {
				prompt.WriteString(" ")
			}

			prompt.WriteString(strings.TrimSpace(message.Content))
			prompt.WriteString(" [/INST]")
		}

		if message.Role == provider.MessageRoleAssistant {
			if i > 0 {
				prompt.WriteString(" ")
			}

			prompt.WriteString(strings.TrimSpace(message.Content))
		}
	}

	return prompt.String(), nil
}

func (t *promptMistral) Stop() []string {
	return []string{
		"[INST]",
		"[/INST]",
	}
}
