package openai

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Provider struct {
	url   string
	token string

	client *openai.Client

	modelMapper ModelMapper
}

type Option func(*Provider)

type ModelMapper = func(model string) string

func New(options ...Option) *Provider {
	p := &Provider{}

	for _, option := range options {
		option(p)
	}

	config := openai.DefaultConfig(p.token)

	if p.url != "" {
		config.BaseURL = p.url
	}

	if strings.Contains(p.url, "openai.azure.com") {
		config = openai.DefaultAzureConfig(p.token, p.url)
	}

	p.client = openai.NewClientWithConfig(config)

	return p
}

func WithURL(url string) Option {
	return func(p *Provider) {
		p.url = url
	}
}

func WithToken(token string) Option {
	return func(p *Provider) {
		p.token = token
	}
}

func WithModelMapper(mapper ModelMapper) Option {
	return func(p *Provider) {
		p.modelMapper = mapper
	}
}

func (p *Provider) Models(ctx context.Context) ([]openai.Model, error) {
	list, err := p.client.ListModels(ctx)

	if err != nil {
		return nil, err
	}

	var result []openai.Model

	for _, model := range list.Models {
		if p.modelMapper != nil {
			id := p.modelMapper(model.ID)

			if id == "" {
				continue
			}

			model.ID = id
		}

		result = append(result, model)
	}

	return result, nil
}

func (p *Provider) Embedding(ctx context.Context, req openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	result, err := p.client.CreateEmbeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	return &result, nil
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
