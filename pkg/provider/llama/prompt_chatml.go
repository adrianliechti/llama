package llama

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type PromptChatML struct {
}

func (t *PromptChatML) Stop() []string {
	return []string{
		"<|im_start",
		"<|im_end",
		"|im_start",
		"|im_end",
	}
}

func (t *PromptChatML) Prompt(system string, messages []provider.Message) (string, error) {
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

	for _, message := range messages {
		if message.Role == provider.MessageRoleSystem {
			prompt.WriteString("<|im_start|>system\n")
			prompt.WriteString(strings.TrimSpace(message.Content))
			prompt.WriteString("<|im_end|>\n")
		}

		if message.Role == provider.MessageRoleUser {
			prompt.WriteString("<|im_start|>user\n")
			prompt.WriteString(strings.TrimSpace(message.Content))
			prompt.WriteString("<|im_end|>\n")
		}

		if message.Role == provider.MessageRoleAssistant {
			prompt.WriteString("<|im_start|>assistant\n")
			prompt.WriteString(strings.TrimSpace(message.Content))
			prompt.WriteString("<|im_end|>\n")

		}
	}

	prompt.WriteString("<|im_start|>assistant")

	return prompt.String(), nil
}
