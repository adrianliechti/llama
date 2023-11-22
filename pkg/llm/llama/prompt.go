package llama

import (
	"errors"
	"slices"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type PromptTemplate interface {
	ConvertMessages(messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error)
	ConvertPrompt(system string, messages []openai.ChatCompletionMessage) (string, error)

	RenderMessages(content string) string
}

type PromptTemplateLLAMA struct {
}

func (t *PromptTemplateLLAMA) ConvertMessages(messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	messages = flattenMessages1(messages)

	if err := verifyMessageOrder1(messages); err != nil {
		return []openai.ChatCompletionMessage{}, err
	}

	return messages, nil
}

func (t *PromptTemplateLLAMA) ConvertPrompt(system string, messages []openai.ChatCompletionMessage) (string, error) {
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

func (t *PromptTemplateLLAMA) RenderMessages(content string) string {
	message := strings.TrimSpace(content)
	return message
}

type PromptTemplateChatML struct {
}

func (t *PromptTemplateChatML) ConvertMessages(messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	messages = flattenMessages1(messages)

	if err := verifyMessageOrder1(messages); err != nil {
		return []openai.ChatCompletionMessage{}, err
	}

	return messages, nil
}

func (t *PromptTemplateChatML) ConvertPrompt(system string, messages []openai.ChatCompletionMessage) (string, error) {
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

func (t *PromptTemplateChatML) RenderMessages(content string) string {
	content = strings.ReplaceAll(content, "<|im_end|>", "")
	return content

}

func flattenMessages1(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, 0)

	for _, m := range messages {
		if len(result) > 0 && result[len(result)-1].Role == m.Role {
			result[len(result)-1].Content += "\n" + m.Content
			continue
		}

		result = append(result, m)
	}

	return result
}

func verifyMessageOrder1(messages []openai.ChatCompletionMessage) error {
	result := slices.Clone(messages)

	if len(result) == 0 {
		return errors.New("there must be at least one message")
	}

	if result[0].Role == openai.ChatMessageRoleSystem {
		result = result[1:]
	}

	errRole := errors.New("model only supports 'system', 'user' and 'assistant' roles, starting with 'system', then 'user' and alternating (u/a/u/a/u...)")

	for i, m := range result {
		if m.Role != openai.ChatMessageRoleUser && m.Role != openai.ChatMessageRoleAssistant {
			return errRole
		}

		if (i+1)%2 == 1 && m.Role != openai.ChatMessageRoleUser {
			return errRole
		}

		if (i+1)%2 == 0 && m.Role != openai.ChatMessageRoleAssistant {
			return errRole
		}
	}

	if result[len(result)-1].Role != openai.ChatMessageRoleUser {
		return errors.New("last message must be from user")
	}

	return nil
}
