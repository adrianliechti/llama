package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/index/chroma"
)

func createIndex(cfg indexConfig) (index.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "chroma":
		return chromaIndex(cfg)

	default:
		return nil, errors.New("invalid index type: " + cfg.Type)
	}
}

func chromaIndex(cfg indexConfig) (index.Provider, error) {
	return chroma.New(cfg.URL, cfg.Name)
}
