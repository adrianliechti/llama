package from

import (
	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/sashabaranov/go-openai"
)

func CompletionMessages(s []provider.CompletionMessage) []openai.ChatCompletionMessage {
	var result []openai.ChatCompletionMessage

	for _, m := range s {
		result = append(result, CompletionMessage(m))
	}

	return result
}

func CompletionMessage(m provider.CompletionMessage) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    MessageRole(m.Role),
		Content: m.Content,
	}
}

func MessageRole(r provider.MessageRole) string {
	switch r {
	case provider.MessageRoleSystem:
		return openai.ChatMessageRoleSystem

	case provider.MessageRoleUser:
		return openai.ChatMessageRoleUser

	case provider.MessageRoleAssistant:
		return openai.ChatMessageRoleAssistant

	default:
		return ""
	}
}

func MessageResult(r provider.MessageResult) openai.FinishReason {
	switch r {
	case "":
		return openai.FinishReasonNull

	case provider.MessageResultStop:
		return openai.FinishReasonStop

	default:
		return openai.FinishReasonStop
	}
}
