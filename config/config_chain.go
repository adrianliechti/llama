package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/chain/rag"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
)

func (c *Config) registerChains(f *configFile) error {
	for id, cfg := range f.Chains {
		var err error

		var index index.Provider

		var embedder provider.Embedder
		var completer provider.Completer

		if cfg.Model != "" {
			completer, err = c.Completer(cfg.Model)

			if err != nil {
				return err
			}
		}

		if cfg.Embedding != "" {
			embedder, err = c.Embedder(cfg.Embedding)

			if err != nil {
				return err
			}
		}

		if cfg.Index != "" {
			index, err = c.Index(cfg.Index)

			if err != nil {
				return err
			}

			if embedder == nil {
				embedder = index
			}
		}

		chain, err := createChain(cfg, embedder, completer, index)

		if err != nil {
			return err
		}

		c.models[id] = Model{
			ID: id,

			model: cfg.Model,
		}

		c.chains[id] = chain
	}

	return nil
}

func createChain(cfg chainConfig, embedder provider.Embedder, completer provider.Completer, index index.Provider) (chain.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "rag":
		return ragChain(cfg, embedder, completer, index)
	default:
		return nil, errors.New("invalid chain type: " + cfg.Type)
	}
}

func ragChain(cfg chainConfig, embedder provider.Embedder, completer provider.Completer, index index.Provider) (chain.Provider, error) {
	var options []rag.Option

	if index != nil {
		options = append(options, rag.WithIndex(index))
	}

	if embedder != nil {
		options = append(options, rag.WithEmbedder(embedder))
	}

	if completer != nil {
		options = append(options, rag.WithCompleter(completer))
	}

	if cfg.TopK != nil {
		options = append(options, rag.WithTopK(*cfg.TopK))
	}

	if cfg.TopP != nil {
		options = append(options, rag.WithTopP(*cfg.TopP))
	}

	return rag.New(options...)
}
