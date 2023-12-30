package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/index/chroma"
)

func createIndex(cfg indexConfig, embedder index.Embedder) (index.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "chroma":
		return chromaIndex(cfg, embedder)

	default:
		return nil, errors.New("invalid index type: " + cfg.Type)
	}
}

func chromaIndex(cfg indexConfig, embedder index.Embedder) (index.Provider, error) {
	var options []chroma.Option

	if embedder != nil {
		options = append(options, chroma.WithEmbedder(embedder))
	}

	return chroma.New(cfg.URL, cfg.Name, options...)
}
