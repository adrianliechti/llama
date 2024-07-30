package provider

import (
	"context"
)

type Completer interface {
	Complete(ctx context.Context, messages []Message, options *CompleteOptions) (*Completion, error)
}

type Message struct {
	Role    MessageRole
	Content string

	Files []File

	Tool      string
	ToolCalls []ToolCall
}

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
)

type ToolCall struct {
	ID string

	Name      string
	Arguments string
}

type CompleteOptions struct {
	Stream chan<- Completion

	Stop  []string
	Tools []Tool

	MaxTokens   *int
	Temperature *float32

	Format CompletionFormat
}

func (o CompleteOptions) Clone() CompleteOptions {
	return CompleteOptions{
		Stream: o.Stream,

		Stop:  o.Stop,
		Tools: o.Tools,

		MaxTokens:   o.MaxTokens,
		Temperature: o.Temperature,

		Format: o.Format,
	}
}

type Completion struct {
	ID string

	Reason CompletionReason

	Message Message
}

type CompletionFormat string

const (
	CompletionFormatJSON CompletionFormat = "json"
)

type CompletionReason string

const (
	CompletionReasonStop     CompletionReason = "stop"
	CompletionReasonLength   CompletionReason = "length"
	CompletionReasonFunction CompletionReason = "function"
)
