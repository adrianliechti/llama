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

type Synthesizer interface {
	Synthesize(ctx context.Context, content string, options *SynthesizeOptions) (*Synthesis, error)
}

type Transcriber interface {
	Transcribe(ctx context.Context, input File, options *TranscribeOptions) (*Transcription, error)
}

type Renderer interface {
	Render(ctx context.Context, input string, options *RenderOptions) (*Image, error)
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

	Stop      []string
	Functions []Function

	MaxTokens   *int
	Temperature *float32

	Format CompletionFormat
}

type Synthesis struct {
	ID string

	Name    string
	Content io.ReadCloser
}

type SynthesizeOptions struct {
	Voice string
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

	Language string
	Duration float64

	Content string
}

type TranscribeOptions struct {
	Language    string
	Temperature *float32
}

type Image struct {
	ID string

	Name    string
	Content io.ReadCloser
}

type RenderOptions struct {
}
