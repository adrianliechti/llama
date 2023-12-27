package provider

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Provider interface {
	Models(ctx context.Context) ([]openai.Model, error)

	Embed(ctx context.Context, req openai.EmbeddingRequest) (*openai.EmbeddingResponse, error)

	Complete(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
	CompleteStream(ctx context.Context, req openai.ChatCompletionRequest, stream chan<- openai.ChatCompletionStreamResponse) error
}
