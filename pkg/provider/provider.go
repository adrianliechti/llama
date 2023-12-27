package provider

import (
	"context"
)

type Provider interface {
	Models(ctx context.Context) ([]Model, error)

	Embed(ctx context.Context, model, content string) (*Embedding, error)

	Complete(ctx context.Context, model string, messages []CompletionMessage) (*Completion, error)
	CompleteStream(ctx context.Context, model string, messages []CompletionMessage, stream chan<- Completion) error
}

type Model struct {
	ID string
}

type Embedding struct {
	Embeddings []float32
}

type CompletionMessage struct {
	Role    MessageRole
	Content string
}

type Completion struct {
	Message CompletionMessage
	Result  MessageResult
}

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type MessageResult string

const (
	MessageResultStop MessageResult = "stop"
)
