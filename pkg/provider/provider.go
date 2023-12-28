package provider

import (
	"context"
)

type Provider interface {
	Models(ctx context.Context) ([]Model, error)

	Embed(ctx context.Context, model, content string) ([]float32, error)

	Complete(ctx context.Context, model string, messages []Message, options *CompleteOptions) (*Message, error)
	CompleteStream(ctx context.Context, model string, messages []Message, stream chan<- Message, options *CompleteOptions) error
}

type Model struct {
	ID string
}

type Message struct {
	Role    MessageRole
	Content string
}

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type CompleteOptions struct {
}
