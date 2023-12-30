package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/index/chroma"
	"github.com/adrianliechti/llama/pkg/index/memory"
)

func (c *Config) registerIndexes(f *configFile) error {
	for id, cfg := range f.Indexes {
		var embedder index.Embedder

		if cfg.Embedding != "" {
			e, err := c.Embedder(cfg.Embedding)

			if err != nil {
				return err
			}

			embedder = e
		}

		i, err := createIndex(cfg, embedder)

		if err != nil {
			return err
		}

		c.indexes[id] = i
	}

	return nil
}

func createIndex(cfg indexConfig, embedder index.Embedder) (index.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "chroma":
		return chromaIndex(cfg, embedder)
	case "memory":
		return memoryIndex(cfg, embedder)

	default:
		return nil, errors.New("invalid index type: " + cfg.Type)
	}
}

func chromaIndex(cfg indexConfig, embedder index.Embedder) (index.Provider, error) {
	var options []chroma.Option

	if embedder != nil {
		options = append(options, chroma.WithEmbedder(embedder))
	}

	return chroma.New(cfg.URL, cfg.Namespace, options...)
}

func memoryIndex(cfg indexConfig, embedder index.Embedder) (index.Provider, error) {
	var options []memory.Option

	if embedder != nil {
		options = append(options, memory.WithEmbedder(embedder))
	}

	return memory.New(options...)
}
