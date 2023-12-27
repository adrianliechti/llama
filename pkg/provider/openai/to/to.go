package to

import (
	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/sashabaranov/go-openai"
)

func CompletionMessages(s []openai.ChatCompletionMessage) []provider.CompletionMessage {
	result := make([]provider.CompletionMessage, 0)

	for i, m := range s {
		result[i] = CompletionMessage(m)
	}

	return result
}

func CompletionMessage(m openai.ChatCompletionMessage) provider.CompletionMessage {
	return provider.CompletionMessage{
		Role:    MessageRole(m.Role),
		Content: m.Content,
	}
}

func MessageRole(r string) provider.MessageRole {
	switch r {
	case openai.ChatMessageRoleSystem:
		return provider.MessageRoleSystem

	case openai.ChatMessageRoleUser:
		return provider.MessageRoleUser

	case openai.ChatMessageRoleAssistant:
		return provider.MessageRoleAssistant

	// case openai.ChatMessageRoleFunction:
	// 	return provider.MessageRoleFunction

	// case openai.ChatMessageRoleTool:
	// 	return provider.MessageRoleTool

	default:
		return provider.MessageRoleAssistant
	}
}

func MessageResult(r openai.FinishReason) provider.MessageResult {
	switch r {
	case openai.FinishReasonNull:
		return ""

	case openai.FinishReasonStop:
		return provider.MessageResultStop

	default:
		return provider.MessageResultStop
	}
}
