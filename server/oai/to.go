package oai

import (
	"github.com/adrianliechti/llama/pkg/provider"
)

func toMessages(s []ChatCompletionMessage) []provider.Message {
	result := make([]provider.Message, 0)

	for _, m := range s {
		result = append(result, toMessage(m))
	}

	return result
}

func toMessage(m ChatCompletionMessage) provider.Message {
	return provider.Message{
		Role:    toMessageRole(m.Role),
		Content: m.Content,
	}
}

func toMessageRole(r MessageRole) provider.MessageRole {
	switch r {
	case MessageRoleSystem:
		return provider.MessageRoleSystem

	case MessageRoleUser:
		return provider.MessageRoleUser

	case MessageRoleAssistant:
		return provider.MessageRoleAssistant

	default:
		return ""
	}
}

func fromMessageRole(r provider.MessageRole) MessageRole {
	switch r {
	case provider.MessageRoleSystem:
		return MessageRoleSystem

	case provider.MessageRoleUser:
		return MessageRoleUser

	case provider.MessageRoleAssistant:
		return MessageRoleAssistant

	default:
		return ""
	}
}

func fromCompletionReason(val provider.CompletionReason) *CompletionReason {
	switch val {
	case provider.CompletionReasonStop:
		return &CompletionReasonStop

	case provider.CompletionReasonLength:
		return &CompletionReasonLength

	default:
		return nil
	}
}
