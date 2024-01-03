package provider

import (
	"context"
)

type Provider interface {
	Embed(ctx context.Context, model, content string) ([]float32, error)
	Complete(ctx context.Context, model string, messages []Message, options *CompleteOptions) (*Completion, error)
}

type Embedder interface {
	Embed(ctx context.Context, content string) ([]float32, error)
}

type Completer interface {
	Complete(ctx context.Context, messages []Message, options *CompleteOptions) (*Completion, error)
}

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleFunction  MessageRole = "function"
)

type Message struct {
	Role    MessageRole
	Content string

	Function      string
	FunctionCalls []FunctionCall
}

type CompletionFormat string

const (
	CompletionFormatJSON CompletionFormat = "json"
)

type Function struct {
	Name       string
	Parameters any

	Description string
}

type FunctionCall struct {
	ID string

	Name      string
	Arguments string
}

type CompletionReason string

const (
	CompletionReasonStop     CompletionReason = "stop"
	CompletionReasonLength   CompletionReason = "length"
	CompletionReasonFunction CompletionReason = "function"
)

type Completion struct {
	ID string

	Reason CompletionReason

	Message Message
}

type CompleteOptions struct {
	Stream chan<- Completion

	Format    CompletionFormat
	Functions []Function

	Stop []string

	Temperature *float32
	TopP        *float32
	MinP        *float32
}

func ToEmbbedder(p Provider, model string) Embedder {
	return &embberder{
		Provider: p,
		model:    model,
	}
}

func ToCompleter(p Provider, model string) Completer {
	return &completer{
		Provider: p,
		model:    model,
	}
}

type embberder struct {
	Provider
	model string
}

func (e *embberder) Embed(ctx context.Context, content string) ([]float32, error) {
	return e.Provider.Embed(ctx, e.model, content)
}

type completer struct {
	Provider
	model string
}

func (c *completer) Complete(ctx context.Context, messages []Message, options *CompleteOptions) (*Completion, error) {
	return c.Provider.Complete(ctx, c.model, messages, options)
}
