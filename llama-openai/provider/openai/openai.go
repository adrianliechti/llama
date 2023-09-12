package openai

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Provider struct {
	client *openai.Client
}

func FromEnvironment() (*Provider, error) {
	token := os.Getenv("OPENAI_API_KEY")

	if token == "" {
		return nil, errors.New("OPENAI_API_KEY is not configured")
	}

	cfg := openai.DefaultConfig(token)

	if val := os.Getenv("OPENAI_API_HOST"); val != "" {
		cfg.BaseURL = val
	}

	client := openai.NewClientWithConfig(cfg)

	return &Provider{
		client: client,
	}, nil
}

func (p *Provider) Models(ctx context.Context) ([]openai.Model, error) {
	result, err := p.client.ListModels(ctx)

	if err != nil {
		return nil, err
	}

	return result.Models, nil
}

func (p *Provider) Chat(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	result, err := p.client.CreateChatCompletion(ctx, req)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *Provider) ChatStream(ctx context.Context, req openai.ChatCompletionRequest, stream chan<- openai.ChatCompletionStreamResponse) error {
	result, err := p.client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		return err
	}

	defer result.Close()

	for {
		response, err := result.Recv()

		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		stream <- response
	}
}
