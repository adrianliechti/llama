package llama

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type PromptChatML struct {
}

func (t *PromptChatML) Stop() []string {
	return []string{
		"<|im_end|>",
	}
}

func (t *PromptChatML) Prompt(system string, messages []provider.Message) (string, error) {
	messages = flattenMessages(messages)

	if err := verifyMessageOrder(messages); err != nil {
		return "", err
	}

	if len(messages) > 0 && messages[0].Role == provider.MessageRoleSystem {
		system = strings.TrimSpace(messages[0].Content)
		messages = messages[1:]
	}

	var prompt string

	for i, message := range messages {
		if message.Role == provider.MessageRoleUser {
			if i == 0 && len(system) > 0 {
				prompt += "<|im_start|>system\n" + strings.TrimSpace(system) + "<|im_end|>\n"
			}

			content := strings.TrimSpace(message.Content)
			prompt += "<|im_start|>user\n" + content + "<|im_end|>\n"
		}

		if message.Role == provider.MessageRoleAssistant {
			content := strings.TrimSpace(message.Content)
			prompt += "<|im_start|>assistant\n" + content + "<|im_end|>\n"
		}
	}

	prompt += "<|im_start|>assistant"

	return strings.TrimSpace(prompt), nil
}
