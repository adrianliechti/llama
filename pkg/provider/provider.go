package provider

import (
	"context"
)

type Provider interface {
	Models(ctx context.Context) ([]Model, error)

	Embedder
	Completer
}

type Embedder interface {
	Embed(ctx context.Context, model, content string) ([]float32, error)
}

type Completer interface {
	Complete(ctx context.Context, model string, messages []Message, options *CompleteOptions) (*Completion, error)
}

type Model struct {
	ID string
}

type Message struct {
	Role    MessageRole
	Content string
}

type CompleteOptions struct {
	Stream chan<- Completion
}

type Completion struct {
	*Message
	Reason CompletionReason
}

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type CompletionReason string

const (
	CompletionReasonStop   CompletionReason = "stop"
	CompletionReasonLength CompletionReason = "length"
)
