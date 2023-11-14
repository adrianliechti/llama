package dispatcher

import (
	"context"
	"errors"

	"github.com/adrianliechti/llama/pkg/llm"

	"github.com/sashabaranov/go-openai"
)

type Provider struct {
	models    map[string]openai.Model
	providers map[string]llm.Provider
}

func New(providers ...llm.Provider) (*Provider, error) {
	p := &Provider{
		models:    map[string]openai.Model{},
		providers: map[string]llm.Provider{},
	}

	for _, provider := range providers {
		models, err := provider.Models(context.Background())

		if err != nil {
			return nil, err
		}

		for _, m := range models {
			p.models[m.ID] = m
			p.providers[m.ID] = provider
		}
	}

	return p, nil
}

func (p *Provider) Models(ctx context.Context) ([]openai.Model, error) {
	result := make([]openai.Model, 0)

	for _, m := range p.models {
		result = append(result, m)
	}

	// sort.Slice(result, func(i, j int) bool {
	// 	return result[i].ID < result[j].ID
	// })

	return result, nil
}

func (p *Provider) Embedding(ctx context.Context, req openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	model := req.Model.String()
	provider, ok := p.providers[model]

	if !ok {
		return nil, errors.New("no provider configured for model")
	}

	return provider.Embedding(ctx, req)
}

func (p *Provider) Chat(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	model := req.Model
	provider, ok := p.providers[model]

	if !ok {
		return nil, errors.New("no provider configured for model")
	}

	return provider.Chat(ctx, req)
}

func (p *Provider) ChatStream(ctx context.Context, req openai.ChatCompletionRequest, stream chan<- openai.ChatCompletionStreamResponse) error {
	model := req.Model
	provider, ok := p.providers[model]

	if !ok {
		return errors.New("no provider configured for model")
	}

	return provider.ChatStream(ctx, req, stream)
}
