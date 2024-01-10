package prompt

import (
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

type promptToRA struct {
}

func (t *promptToRA) Prompt(system string, messages []provider.Message, options *TemplateOptions) (string, error) {
	if options == nil {
		options = new(TemplateOptions)
	}

	if len(options.Functions) > 0 {
		return "", ErrFunctionsUnsupported
	}

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
			prompt.WriteString("<|user|>\n")

			if i == 0 && len(system) > 0 {
				prompt.WriteString(strings.TrimSpace(system))
				prompt.WriteString("\n\n")
			}

			prompt.WriteString(strings.TrimSpace(message.Content))
			prompt.WriteString("\n")
		}

		if message.Role == provider.MessageRoleAssistant {
			prompt.WriteString("<|assistant|>\n")
			prompt.WriteString(strings.TrimSpace(message.Content))
			prompt.WriteString("\n")
		}
	}

	prompt.WriteString("<|assistant|>\n")

	return prompt.String(), nil
}

func (t *promptToRA) Stop() []string {
	return []string{
		"<|user|>",
		"<|assistant|>",
	}
}
