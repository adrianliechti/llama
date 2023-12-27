package llama

import (
	"strings"

	"github.com/sashabaranov/go-openai"
)

type PromptTemplateLLAMA struct {
}

func (t *PromptTemplateLLAMA) ConvertPrompt(system string, messages []openai.ChatCompletionMessage) (string, error) {
	messages = flattenMessages(messages)

	if err := verifyMessageOrder(messages); err != nil {
		return "", err
	}

	if len(messages) > 0 && messages[0].Role == openai.ChatMessageRoleSystem {
		system = strings.TrimSpace(messages[0].Content)
		messages = messages[1:]
	}

	var prompt string

	for i, message := range messages {
		if message.Role == openai.ChatMessageRoleUser {
			content := strings.TrimSpace(message.Content)

			if i == 0 && len(system) > 0 {
				content = "<<SYS>>\n" + system + "\n<</SYS>>\n\n" + content
			}

			prompt += " [INST] " + content + " [/INST]"
		}

		if message.Role == openai.ChatMessageRoleAssistant {
			content := strings.TrimSpace(message.Content)
			prompt += " " + content
		}
	}

	return strings.TrimSpace(prompt), nil
}

func (t *PromptTemplateLLAMA) RenderContent(content string) string {
	message := strings.TrimSpace(content)
	return message
}
