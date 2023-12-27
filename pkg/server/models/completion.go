package models

type ChatCompletionRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream,omitempty"`

	Messages []ChatCompletionMessage `json:"messages"`
}

type ChatCompletion struct {
	ID string `json:"id"`

	Object  string `json:"object"`
	Created int64  `json:"created"`

	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
}

type ChatCompletionChoice struct {
	Index int `json:"index"`

	Delta   ChatCompletionMessage `json:"delta,omitempty"`
	Message ChatCompletionMessage `json:"message,omitempty"`

	FinishReason FinishReason `json:"finish_reason,omitempty"`
}

type ChatCompletionMessage struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
}

type MessageRole string

const (
	MessageRoleSystem    = "system"
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
)

type FinishReason string

const (
	FinishReasonStop   FinishReason = "stop"
	FinishReasonLength FinishReason = "length"
)
