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

type StreamHandler = func(ctx context.Context, completion Completion) error

type CompleteOptions struct {
	Stream StreamHandler

	Stop  []string
	Tools []Tool

	MaxTokens   *int
	Temperature *float32

	Format CompletionFormat
	Schema *Schema
}

type Completion struct {
	ID string

	Reason CompletionReason

	Message Message

	Usage *Usage
}

type CompletionFormat string

const (
	CompletionFormatJSON CompletionFormat = "json"
)

type CompletionReason string

const (
	CompletionReasonStop   CompletionReason = "stop"
	CompletionReasonLength CompletionReason = "length"
	CompletionReasonTool   CompletionReason = "tool"
	CompletionReasonFilter CompletionReason = "filter"
)
