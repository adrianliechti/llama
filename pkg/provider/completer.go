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

	Function      string
	FunctionCalls []FunctionCall
}

type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleFunction  MessageRole = "function"
)

type FunctionCall struct {
	ID string

	Name      string
	Arguments string
}

type CompleteOptions struct {
	Stream chan<- Completion

	Stop      []string
	Functions []Function

	MaxTokens   *int
	Temperature *float32

	Format CompletionFormat
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
