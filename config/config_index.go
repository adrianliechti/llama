package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/index/chroma"
	"github.com/adrianliechti/llama/pkg/index/elasticsearch"
	"github.com/adrianliechti/llama/pkg/index/memory"
	"github.com/adrianliechti/llama/pkg/index/weaviate"
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

	case "weaviate":
		return weaviateIndex(cfg, embedder)

	case "elasticsearch":
		return elasticsearchIndex(cfg)

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

func weaviateIndex(cfg indexConfig, embedder index.Embedder) (index.Provider, error) {
	var options []weaviate.Option

	if embedder != nil {
		options = append(options, weaviate.WithEmbedder(embedder))
	}

	return weaviate.New(cfg.URL, cfg.Namespace, options...)
}

func elasticsearchIndex(cfg indexConfig) (index.Provider, error) {
	var options []elasticsearch.Option

	return elasticsearch.New(cfg.URL, cfg.Namespace, options...)
}
