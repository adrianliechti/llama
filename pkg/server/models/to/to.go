package to

import (
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/server/models"
)

func CompletionMessages(s []models.ChatCompletionMessage) []provider.CompletionMessage {
	result := make([]provider.CompletionMessage, 0)

	for _, m := range s {
		result = append(result, CompletionMessage(m))
	}

	return result
}

func CompletionMessage(m models.ChatCompletionMessage) provider.CompletionMessage {
	return provider.CompletionMessage{
		Role:    MessageRole(m.Role),
		Content: m.Content,
	}
}

func MessageRole(r models.MessageRole) provider.MessageRole {
	switch r {
	case models.MessageRoleSystem:
		return provider.MessageRoleSystem

	case models.MessageRoleUser:
		return provider.MessageRoleUser

	case models.MessageRoleAssistant:
		return provider.MessageRoleAssistant

	default:
		return ""
	}
}

func MessageResult(r models.FinishReason) provider.MessageResult {
	switch r {
	case "":
		return ""

	case models.FinishReasonStop:
		return provider.MessageResultStop

	case models.FinishReasonLength:
		return provider.MessageResultLength

	default:
		return provider.MessageResultStop
	}
}
