package openai

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var (
	ErrInvalidModelMapping = errors.New("invalid model mapping")
)

type Provider struct {
	url   string
	token string

	client *openai.Client
	mapper ModelMapper
}

type ModelMapper interface {
	From(key string) string
	To(key string) string
}

type Option func(*Provider)

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
		p.mapper = mapper
	}
}

func (p *Provider) Models(ctx context.Context) ([]openai.Model, error) {
	list, err := p.client.ListModels(ctx)

	if err != nil {
		return nil, err
	}

	if p.mapper == nil {
		return list.Models, nil
	}

	var result []openai.Model

	for _, m := range list.Models {
		model := p.mapper.From(m.ID)

		if model == "" {
			continue
		}

		m.ID = model

		result = append(result, m)
	}

	return result, nil
}

func (p *Provider) Embed(ctx context.Context, req openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	if p.mapper != nil {
		model := p.mapper.To(req.Model.String())

		if model == "" {
			return nil, ErrInvalidModelMapping
		}

		req.Model = openai.Unknown

		if strings.EqualFold(model, "text-embedding-ada-002") {
			req.Model = openai.AdaEmbeddingV2
		}
	}

	result, err := p.client.CreateEmbeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *Provider) Complete(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	if p.mapper != nil {
		model := p.mapper.To(req.Model)

		if model == "" {
			return nil, ErrInvalidModelMapping
		}

		req.Model = model
	}

	result, err := p.client.CreateChatCompletion(ctx, req)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *Provider) CompleteStream(ctx context.Context, req openai.ChatCompletionRequest, stream chan<- openai.ChatCompletionStreamResponse) error {
	if p.mapper != nil {
		model := p.mapper.To(req.Model)

		if model == "" {
			return ErrInvalidModelMapping
		}

		req.Model = model
	}

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
