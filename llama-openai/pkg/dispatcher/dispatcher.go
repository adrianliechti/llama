package dispatcher

import (
	"context"
	"errors"

	"chat/provider"

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

func (p Provider) Models(ctx context.Context) ([]openai.Model, error) {
	result := make([]openai.Model, 0)

	for _, m := range p.models {
		result = append(result, m)
	}

	return result, nil
}

func (p *Provider) Chat(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	provider, ok := p.providers[request.Model]

	if !ok {
		return nil, errors.New("no provider configured for model")
	}

	return provider.Chat(ctx, request)
}

func (p *Provider) ChatStream(ctx context.Context, request openai.ChatCompletionRequest, stream chan<- openai.ChatCompletionStreamResponse) error {
	provider, ok := p.providers[request.Model]

	if !ok {
		return errors.New("no provider configured for model")
	}

	return provider.ChatStream(ctx, request, stream)
}
