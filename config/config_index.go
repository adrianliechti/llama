package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/index/azure"
	"github.com/adrianliechti/llama/pkg/index/chroma"
	"github.com/adrianliechti/llama/pkg/index/custom"
	"github.com/adrianliechti/llama/pkg/index/elasticsearch"
	"github.com/adrianliechti/llama/pkg/index/memory"
	"github.com/adrianliechti/llama/pkg/index/qdrant"
	"github.com/adrianliechti/llama/pkg/index/weaviate"
	"github.com/adrianliechti/llama/pkg/otel"
)

func (cfg *Config) RegisterIndex(name, alias string, p index.Provider) {
	if cfg.indexes == nil {
		cfg.indexes = make(map[string]index.Provider)
	}

	cfg.indexes[alias] = p
}

func (cfg *Config) Index(id string) (index.Provider, error) {
	if cfg.indexes != nil {
		if i, ok := cfg.indexes[id]; ok {
			return i, nil
		}
	}

	return nil, errors.New("index not found: " + id)
}

type indexConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Namespace string `yaml:"namespace"`

	Embedder string `yaml:"embedder"`
	Reranker string `yaml:"reranker"`
}

type indexContext struct {
	Embedder index.Embedder
	Reranker index.Reranker
}

func (cfg *Config) registerIndexes(f *configFile) error {
	for id, i := range f.Indexes {
		var err error
		context := indexContext{}

		if i.Embedder != "" {
			if context.Embedder, err = cfg.Embedder(i.Embedder); err != nil {
				return err
			}
		}

		if i.Reranker != "" {
			if context.Reranker, err = cfg.Reranker(i.Reranker); err != nil {
				return err
			}
		}

		index, err := createIndex(i, context)

		if err != nil {
			return err
		}

		if _, ok := index.(otel.Index); !ok {
			index = otel.NewIndex(i.Type, id, index)
		}

		cfg.RegisterIndex(i.Type, id, index)
	}

	return nil
}

func createIndex(cfg indexConfig, context indexContext) (index.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "azure":
		return azureIndex(cfg)

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

func azureIndex(cfg indexConfig) (index.Provider, error) {
	var options []azure.Option

	return azure.New(cfg.URL, cfg.Namespace, cfg.Token, options...)
}

func chromaIndex(cfg indexConfig, context indexContext) (index.Provider, error) {
	var options []chroma.Option

	if context.Embedder != nil {
		options = append(options, chroma.WithEmbedder(context.Embedder))
	}

	if context.Reranker != nil {
		options = append(options, chroma.WithReranker(context.Reranker))
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

	if context.Reranker != nil {
		options = append(options, memory.WithReranker(context.Reranker))
	}

	return memory.New(options...)
}

func qdrantIndex(cfg indexConfig, context indexContext) (index.Provider, error) {
	var options []qdrant.Option

	if context.Embedder != nil {
		options = append(options, qdrant.WithEmbedder(context.Embedder))
	}

	if context.Reranker != nil {
		options = append(options, qdrant.WithReranker(context.Reranker))
	}

	return qdrant.New(cfg.URL, cfg.Namespace, options...)
}

func weaviateIndex(cfg indexConfig, context indexContext) (index.Provider, error) {
	var options []weaviate.Option

	if context.Embedder != nil {
		options = append(options, weaviate.WithEmbedder(context.Embedder))
	}

	if context.Reranker != nil {
		options = append(options, weaviate.WithReranker(context.Reranker))
	}

	return weaviate.New(cfg.URL, cfg.Namespace, options...)
}

func customIndex(cfg indexConfig) (*custom.Client, error) {
	var options []custom.Option

	return custom.New(cfg.URL, options...)
}
