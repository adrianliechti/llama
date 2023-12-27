package openai

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/openai/from"
	"github.com/adrianliechti/llama/pkg/provider/openai/to"
	"github.com/sashabaranov/go-openai"
)

var (
	_ provider.Provider = &Provider{}
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

func (p *Provider) Models(ctx context.Context) ([]provider.Model, error) {
	list, err := p.client.ListModels(ctx)

	if err != nil {
		return nil, err
	}

	var result []provider.Model

	for _, m := range list.Models {
		model := provider.Model{
			ID: m.ID,
		}

		if p.mapper != nil {
			model.ID = p.mapper.From(m.ID)
		}

		if model.ID == "" {
			continue
		}

		result = append(result, model)
	}

	return result, nil
}

func (p *Provider) Embed(ctx context.Context, model, content string) (*provider.Embedding, error) {
	// if p.mapper != nil {
	// 	model = p.mapper.To(model)
	// }

	req := openai.EmbeddingRequest{
		Input: content,
		Model: openai.AdaEmbeddingV2,
	}

	data, err := p.client.CreateEmbeddings(ctx, req)

	if err != nil {
		return nil, err
	}

	result := provider.Embedding{
		Embeddings: data.Data[0].Embedding,
	}

	return &result, nil
}

func (p *Provider) Complete(ctx context.Context, model string, messages []provider.CompletionMessage) (*provider.Completion, error) {
	if p.mapper != nil {
		model = p.mapper.To(model)
	}

	if model == "" {
		return nil, ErrInvalidModelMapping
	}

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: from.CompletionMessages(messages),
	}

	result, err := p.client.CreateChatCompletion(ctx, req)

	if err != nil {
		return nil, err
	}

	completion := provider.Completion{
		Message: provider.CompletionMessage{
			Role:    to.MessageRole(result.Choices[0].Message.Role),
			Content: result.Choices[0].Message.Content,
		},

		Result: provider.MessageResultStop,
	}

	return &completion, nil
}

func (p *Provider) CompleteStream(ctx context.Context, model string, messages []provider.CompletionMessage, stream chan<- provider.Completion) error {
	if p.mapper != nil {
		model = p.mapper.To(model)
	}

	if model == "" {
		return ErrInvalidModelMapping
	}

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: from.CompletionMessages(messages),
	}

	result, err := p.client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		return err
	}

	defer result.Close()

	for {
		data, err := result.Recv()

		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		completion := provider.Completion{
			Message: provider.CompletionMessage{
				Role:    to.MessageRole(data.Choices[0].Delta.Role),
				Content: data.Choices[0].Delta.Content,
			},

			Result: to.MessageResult(data.Choices[0].FinishReason),
		}

		stream <- completion
	}
}
