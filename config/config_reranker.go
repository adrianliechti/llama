package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/huggingface"
	"github.com/adrianliechti/llama/pkg/provider/jina"
)

func (cfg *Config) RegisterReranker(model string, p provider.Reranker) {
	cfg.RegisterModel(model)

	if cfg.reranker == nil {
		cfg.reranker = make(map[string]provider.Reranker)
	}

	cfg.reranker[model] = p
}

func (cfg *Config) Reranker(model string) (provider.Reranker, error) {
	if cfg.reranker != nil {
		if e, ok := cfg.reranker[model]; ok {
			return e, nil
		}
	}

	return nil, errors.New("reranker not found: " + model)
}

func createReranker(cfg providerConfig, model modelContext) (provider.Reranker, error) {
	switch strings.ToLower(cfg.Type) {
	case "huggingface":
		return huggingfaceReranker(cfg, model)

	case "jina":
		return jinaReranker(cfg, model)

	default:
		return nil, errors.New("invalid reranker type: " + cfg.Type)
	}
}

func huggingfaceReranker(cfg providerConfig, model modelContext) (provider.Reranker, error) {
	var options []huggingface.Option

	if cfg.Token != "" {
		options = append(options, huggingface.WithToken(cfg.Token))
	}

	return huggingface.NewReranker(cfg.URL, model.ID, options...)
}

func jinaReranker(cfg providerConfig, model modelContext) (provider.Reranker, error) {
	var options []jina.Option

	if cfg.Token != "" {
		options = append(options, jina.WithToken(cfg.Token))
	}

	return jina.NewReranker(cfg.URL, model.ID, options...)
}
