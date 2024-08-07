package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/openai"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (cfg *Config) RegisterRenderer(name, model string, p provider.Renderer) {
	cfg.RegisterModel(model)

	if cfg.renderer == nil {
		cfg.renderer = make(map[string]provider.Renderer)
	}

	renderer, ok := p.(otel.ObservableRenderer)

	if !ok {
		renderer = otel.NewRenderer(name, model, p)
	}

	cfg.renderer[model] = renderer
}

func (cfg *Config) Renderer(model string) (provider.Renderer, error) {
	if cfg.renderer != nil {
		if t, ok := cfg.renderer[model]; ok {
			return t, nil
		}
	}

	return nil, errors.New("renderer not found: " + model)
}

func createRenderer(cfg providerConfig, model modelContext) (provider.Renderer, error) {
	switch strings.ToLower(cfg.Type) {
	case "openai":
		return openaiRenderer(cfg, model)

	default:
		return nil, errors.New("invalid renderer type: " + cfg.Type)
	}
}

func openaiRenderer(cfg providerConfig, model modelContext) (provider.Renderer, error) {
	var options []openai.Option

	if cfg.URL != "" {
		options = append(options, openai.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, openai.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, openai.WithModel(model.ID))
	}

	return openai.NewRenderer(options...)
}
