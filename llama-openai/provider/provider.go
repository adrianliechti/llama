package provider

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Provider interface {
	Models(ctx context.Context) ([]openai.Model, error)

	Chat(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
	ChatStream(ctx context.Context, req openai.ChatCompletionRequest, stream chan<- openai.ChatCompletionStreamResponse) error
}
