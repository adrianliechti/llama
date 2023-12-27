package llama

import (
	"strings"

	"github.com/sashabaranov/go-openai"
)

type PromptTemplateChatML struct {
}

func (t *PromptTemplateChatML) ConvertPrompt(system string, messages []openai.ChatCompletionMessage) (string, error) {
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
			if i == 0 && len(system) > 0 {
				prompt += "<|im_start|>system\n" + strings.TrimSpace(system) + "<|im_end|>\n"
			}

			content := strings.TrimSpace(message.Content)
			prompt += "<|im_start|>user\n" + content + "<|im_end|>\n"
		}

		if message.Role == openai.ChatMessageRoleAssistant {
			content := strings.TrimSpace(message.Content)
			prompt += "<|im_start|>assistant\n" + content + "<|im_end|>\n"
		}
	}

	prompt += "<|im_start|>assistant"

	return strings.TrimSpace(prompt), nil
}

func (t *PromptTemplateChatML) RenderContent(content string) string {
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "<|im_end|>", "")
	return content

}
