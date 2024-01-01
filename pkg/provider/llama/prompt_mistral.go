package llama

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type PromptMistral struct {
}

func (t *PromptMistral) Stop() []string {
	return []string{
		"[INST]",
		"[/INST]",
	}
}

func (t *PromptMistral) Prompt(system string, messages []provider.Message) (string, error) {
	messages = llamaMessageFlattening(messages)

	if err := llamaMessageOrder(messages); err != nil {
		return "", err
	}

	if len(messages) > 0 && messages[0].Role == provider.MessageRoleSystem {
		system = strings.TrimSpace(messages[0].Content)
		messages = messages[1:]
	}

	var prompt string

	for i, message := range messages {
		if message.Role == provider.MessageRoleUser {
			content := strings.TrimSpace(message.Content)

			if i == 0 && len(system) > 0 {
				content = system + "\n\n" + content
			}

			prompt += " [INST] " + content + " [/INST]"
		}

		if message.Role == provider.MessageRoleAssistant {
			content := strings.TrimSpace(message.Content)
			prompt += " " + content
		}
	}

	return strings.TrimSpace(prompt), nil
}
