package dispatcher

import (
	"context"
	"errors"
	"sort"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/sashabaranov/go-openai"
)

type Provider struct {
	models    map[string]openai.Model
	providers map[string]provider.Provider
}

func New(providers ...provider.Provider) (*Provider, error) {
	p := &Provider{
		models:    map[string]openai.Model{},
		providers: map[string]provider.Provider{},
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

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

func (p *Provider) Embed(ctx context.Context, req openai.EmbeddingRequest) (*openai.EmbeddingResponse, error) {
	model := req.Model.String()
	provider, ok := p.providers[model]

	if !ok {
		return nil, errors.New("no provider configured for model")
	}

	return provider.Embed(ctx, req)
}

func (p *Provider) Complete(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	model := req.Model
	provider, ok := p.providers[model]

	if !ok {
		return nil, errors.New("no provider configured for model")
	}

	return provider.Complete(ctx, req)
}

func (p *Provider) CompleteStream(ctx context.Context, req openai.ChatCompletionRequest, stream chan<- openai.ChatCompletionStreamResponse) error {
	model := req.Model
	provider, ok := p.providers[model]

	if !ok {
		return errors.New("no provider configured for model")
	}

	return provider.CompleteStream(ctx, req, stream)
}
