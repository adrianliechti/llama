package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/cohere"
	"github.com/adrianliechti/llama/pkg/provider/huggingface"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

func (cfg *Config) RegisterEmbedder(model string, e provider.Embedder) {
	cfg.RegisterModel(model)

	if cfg.embedder == nil {
		cfg.embedder = make(map[string]provider.Embedder)
	}

	cfg.embedder[model] = e
}

func createEmbedder(cfg providerConfig, model string) (provider.Embedder, error) {
	switch strings.ToLower(cfg.Type) {

	case "cohere":
		return cohereEmbedder(cfg, model)

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

func cohereEmbedder(cfg providerConfig, model string) (provider.Embedder, error) {
	var options []cohere.Option

	if cfg.Token != "" {
		options = append(options, cohere.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, cohere.WithModel(model))
	}

	return cohere.NewEmbedder(options...)
}

func huggingfaceEmbedder(cfg providerConfig, model string) (provider.Embedder, error) {
	var options []huggingface.Option

	if cfg.Token != "" {
		options = append(options, huggingface.WithToken(cfg.Token))
	}

	return huggingface.NewEmbedder(cfg.URL, options...)
}

func llamaEmbedder(cfg providerConfig, model string) (provider.Embedder, error) {
	var options []llama.Option

	if model != "" {
		options = append(options, llama.WithModel(model))
	}

	return llama.NewEmbedder(cfg.URL, options...)
}

func ollamaEmbedder(cfg providerConfig, model string) (provider.Embedder, error) {
	var options []ollama.Option

	if model != "" {
		options = append(options, ollama.WithModel(model))
	}

	return ollama.NewEmbedder(cfg.URL, options...)
}

func openaiEmbedder(cfg providerConfig, model string) (provider.Embedder, error) {
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

	return openai.NewEmbedder(options...)
}
