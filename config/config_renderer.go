package config

import (
	"errors"
	"slices"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/openai"
	"github.com/adrianliechti/llama/pkg/provider/replicate/flux"
)

func (cfg *Config) RegisterRenderer(model string, p provider.Renderer) {
	cfg.RegisterModel(model)

	if cfg.renderer == nil {
		cfg.renderer = make(map[string]provider.Renderer)
	}

	cfg.renderer[model] = p
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

	case "replicate":
		return replicateRenderer(cfg, model)

	default:
		return nil, errors.New("invalid renderer type: " + cfg.Type)
	}
}

func openaiRenderer(cfg providerConfig, model modelContext) (provider.Renderer, error) {
	var options []openai.Option

	if cfg.Token != "" {
		options = append(options, openai.WithToken(cfg.Token))
	}

	return openai.NewRenderer(cfg.URL, model.ID, options...)
}

func replicateRenderer(cfg providerConfig, model modelContext) (provider.Renderer, error) {
	if slices.Contains(flux.SupportedModels, model.ID) {
		var options []flux.Option

		if cfg.Token != "" {
			options = append(options, flux.WithToken(cfg.Token))
		}

		return flux.NewRenderer(cfg.URL, model.ID, options...)
	}

	return nil, errors.New("model not supported: " + model.ID)
}
