package provider

import (
	"context"
	"io"
)

type Embedder interface {
	Embed(ctx context.Context, content string) ([]float32, error)
}

type Completer interface {
	Complete(ctx context.Context, messages []Message, options *CompleteOptions) (*Completion, error)
}

type Transcriber interface {
	Transcribe(ctx context.Context, input File, options *TranscribeOptions) (*Transcription, error)
}

type Model struct {
	ID string
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

	Files []File

	Function      string
	FunctionCalls []FunctionCall
}

type CompletionFormat string

const (
	CompletionFormatJSON CompletionFormat = "json"
)

type File struct {
	ID string

	Name    string
	Content io.Reader
}

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

	Format CompletionFormat

	Functions []Function

	Stop []string

	Temperature *float32
	TopP        *float32
	MinP        *float32
}

type Transcription struct {
	ID string

	Content string
}

type TranscribeOptions struct {
	Language    string
	Temperature *float32
}
