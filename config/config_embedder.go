package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/azureai"
	"github.com/adrianliechti/llama/pkg/provider/cohere"
	"github.com/adrianliechti/llama/pkg/provider/huggingface"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
	"github.com/adrianliechti/llama/pkg/provider/openai"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (cfg *Config) RegisterEmbedder(name, model string, p provider.Embedder) {
	cfg.RegisterModel(model)

	if cfg.embedder == nil {
		cfg.embedder = make(map[string]provider.Embedder)
	}

	embedder, ok := p.(otel.ObservableEmbedder)

	if !ok {
		embedder = otel.NewEmbedder(name, model, p)
	}

	cfg.embedder[model] = embedder
}

func (cfg *Config) Embedder(model string) (provider.Embedder, error) {
	if cfg.embedder != nil {
		if e, ok := cfg.embedder[model]; ok {
			return e, nil
		}
	}

	return nil, errors.New("embedder not found: " + model)
}

func createEmbedder(cfg providerConfig, model modelContext) (provider.Embedder, error) {
	switch strings.ToLower(cfg.Type) {
	case "azureai":
		return azureaiEmbedder(cfg, model)

	case "cohere":
		return cohereEmbedder(cfg, model)

	case "github":
		return azureaiEmbedder(cfg, model)

	case "huggingface":
		return huggingfaceEmbedder(cfg, model)

	case "llama":
		return llamaEmbedder(cfg, model)

	case "ollama":
		return ollamaEmbedder(cfg, model)

	case "openai":
		return openaiEmbedder(cfg, model)

	default:
		return nil, errors.New("invalid embedder type: " + cfg.Type)
	}
}

func azureaiEmbedder(cfg providerConfig, model modelContext) (provider.Embedder, error) {
	var options []azureai.Option

	if cfg.URL != "" {
		options = append(options, azureai.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, azureai.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, azureai.WithModel(model.ID))
	}

	return azureai.NewEmbedder(cfg.URL, options...)
}

func cohereEmbedder(cfg providerConfig, model modelContext) (provider.Embedder, error) {
	var options []cohere.Option

	if cfg.Token != "" {
		options = append(options, cohere.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, cohere.WithModel(model.ID))
	}

	return cohere.NewEmbedder(options...)
}

func huggingfaceEmbedder(cfg providerConfig, model modelContext) (provider.Embedder, error) {
	var options []huggingface.Option

	if cfg.Token != "" {
		options = append(options, huggingface.WithToken(cfg.Token))
	}

	return huggingface.NewEmbedder(cfg.URL, options...)
}

func llamaEmbedder(cfg providerConfig, model modelContext) (provider.Embedder, error) {
	var options []llama.Option

	if model.ID != "" {
		options = append(options, llama.WithModel(model.ID))
	}

	return llama.NewEmbedder(cfg.URL, options...)
}

func ollamaEmbedder(cfg providerConfig, model modelContext) (provider.Embedder, error) {
	var options []ollama.Option

	if model.ID != "" {
		options = append(options, ollama.WithModel(model.ID))
	}

	return ollama.NewEmbedder(cfg.URL, options...)
}

func openaiEmbedder(cfg providerConfig, model modelContext) (provider.Embedder, error) {
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

	return openai.NewEmbedder(options...)
}
