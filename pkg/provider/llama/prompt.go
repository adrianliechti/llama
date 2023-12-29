package llama

import (
	"errors"
	"slices"

	"github.com/adrianliechti/llama/pkg/provider"
)

type PromptTemplate interface {
	Stop() []string
	Prompt(system string, messages []provider.Message) (string, error)
}

func flattenMessages(messages []provider.Message) []provider.Message {
	result := make([]provider.Message, 0)

	for _, m := range messages {
		if len(result) > 0 && result[len(result)-1].Role == m.Role {
			result[len(result)-1].Content += "\n" + m.Content
			continue
		}

		result = append(result, m)
	}

	return result
}

func verifyMessageOrder(messages []provider.Message) error {
	result := slices.Clone(messages)

	if len(result) == 0 {
		return errors.New("there must be at least one message")
	}

	if result[0].Role == provider.MessageRoleSystem {
		result = result[1:]
	}

	errRole := errors.New("model only supports 'system', 'user' and 'assistant' roles, starting with 'system', then 'user' and alternating (u/a/u/a/u...)")

	for i, m := range result {
		if m.Role != provider.MessageRoleUser && m.Role != provider.MessageRoleAssistant {
			return errRole
		}

		if (i+1)%2 == 1 && m.Role != provider.MessageRoleUser {
			return errRole
		}

		if (i+1)%2 == 0 && m.Role != provider.MessageRoleAssistant {
			return errRole
		}
	}

	if result[len(result)-1].Role != provider.MessageRoleUser {
		return errors.New("last message must be from user")
	}

	return nil
}
