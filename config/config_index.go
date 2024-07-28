package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/index/aisearch"
	"github.com/adrianliechti/llama/pkg/index/chroma"
	"github.com/adrianliechti/llama/pkg/index/custom"
	"github.com/adrianliechti/llama/pkg/index/elasticsearch"
	"github.com/adrianliechti/llama/pkg/index/memory"
	"github.com/adrianliechti/llama/pkg/index/qdrant"
	"github.com/adrianliechti/llama/pkg/index/weaviate"
)

func (cfg *Config) RegisterIndex(id string, i index.Provider) {
	if cfg.indexes == nil {
		cfg.indexes = make(map[string]index.Provider)
	}

	cfg.indexes[id] = i
}

type indexContext struct {
	Embedder index.Embedder
}

func (cfg *Config) registerIndexes(f *configFile) error {
	for id, i := range f.Indexes {
		var err error
		context := indexContext{}

		if i.Embedding != "" {
			if context.Embedder, err = cfg.Embedder(i.Embedding); err != nil {
				return err
			}
		}

		index, err := createIndex(i, context)

		if err != nil {
			return err
		}

		cfg.RegisterIndex(id, index)
	}

	return nil
}

func createIndex(cfg indexConfig, context indexContext) (index.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "aisearch":
		return aisearchIndex(cfg)

	case "chroma":
		return chromaIndex(cfg, context)

	case "elasticsearch":
		return elasticsearchIndex(cfg)

	case "memory":
		return memoryIndex(cfg, context)

	case "qdrant":
		return qdrantIndex(cfg, context)

	case "weaviate":
		return weaviateIndex(cfg, context)

	case "custom":
		return customIndex(cfg)

	default:
		return nil, errors.New("invalid index type: " + cfg.Type)
	}
}

func aisearchIndex(cfg indexConfig) (index.Provider, error) {
	var options []aisearch.Option

	return aisearch.New(cfg.URL, cfg.Namespace, cfg.Token, options...)
}

func chromaIndex(cfg indexConfig, context indexContext) (index.Provider, error) {
	var options []chroma.Option

	if context.Embedder != nil {
		options = append(options, chroma.WithEmbedder(context.Embedder))
	}

	return chroma.New(cfg.URL, cfg.Namespace, options...)
}

func elasticsearchIndex(cfg indexConfig) (index.Provider, error) {
	var options []elasticsearch.Option

	return elasticsearch.New(cfg.URL, cfg.Namespace, options...)
}

func memoryIndex(cfg indexConfig, context indexContext) (index.Provider, error) {
	var options []memory.Option

	if context.Embedder != nil {
		options = append(options, memory.WithEmbedder(context.Embedder))
	}

	return memory.New(options...)
}

func qdrantIndex(cfg indexConfig, context indexContext) (index.Provider, error) {
	var options []qdrant.Option

	if context.Embedder != nil {
		options = append(options, qdrant.WithEmbedder(context.Embedder))
	}

	return qdrant.New(cfg.URL, cfg.Namespace, options...)
}

func weaviateIndex(cfg indexConfig, context indexContext) (index.Provider, error) {
	var options []weaviate.Option

	if context.Embedder != nil {
		options = append(options, weaviate.WithEmbedder(context.Embedder))
	}

	return weaviate.New(cfg.URL, cfg.Namespace, options...)
}

func customIndex(cfg indexConfig) (*custom.Client, error) {
	var options []custom.Option

	return custom.New(cfg.URL, options...)
}
