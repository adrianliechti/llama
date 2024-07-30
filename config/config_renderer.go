package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/otel"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/openai"
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

func createRenderer(cfg providerConfig, model string) (provider.Renderer, error) {
	switch strings.ToLower(cfg.Type) {

	case "openai":
		return openaiRenderer(cfg, model)

	default:
		return nil, errors.New("invalid renderer type: " + cfg.Type)
	}
}

func openaiRenderer(cfg providerConfig, model string) (provider.Renderer, error) {
	var options []openai.Option

	if cfg.URL != "" {
		options = append(options, openai.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, openai.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, openai.WithModel(model))
	}

	return openai.NewRenderer(options...)
}
