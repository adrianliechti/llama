package provider

import (
	"context"
	"io"

	"github.com/adrianliechti/llama/pkg/jsonschema"
)

type Embedder interface {
	Embed(ctx context.Context, content string) ([]float32, error)
}

type Completer interface {
	Complete(ctx context.Context, messages []Message, options *CompleteOptions) (*Completion, error)
}

type Translator interface {
	Translate(ctx context.Context, content string, options *TranslateOptions) (*Translation, error)
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
	Name        string
	Description string

	Parameters jsonschema.Definition
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
}

type Translation struct {
	ID string

	Content string
}

type TranslateOptions struct {
	Language string
}

type Transcription struct {
	ID string

	Content string
}

type TranscribeOptions struct {
	Language    string
	Temperature *float32
}
