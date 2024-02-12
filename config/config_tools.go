package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/adrianliechti/llama/pkg/tool/search"
)

func (c *Config) registerTools(f *configFile) error {
	for id, cfg := range f.Tools {
		var err error

		var index index.Provider
		var embedder provider.Embedder
		var completer provider.Completer

		if cfg.Index != "" {
			index, err = c.Index(cfg.Index)

			if err != nil {
				return err
			}
		}

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

		t, err := createTool(cfg, embedder, completer, index)

		if err != nil {
			return err
		}

		c.tools[id] = t
	}

	return nil
}

func createTool(cfg toolConfig, embedder provider.Embedder, completer provider.Completer, index index.Provider) (tool.Tool, error) {
	switch strings.ToLower(cfg.Type) {
	case "search":
		return searchTool(cfg, index)

	default:
		return nil, errors.New("invalid tool type: " + cfg.Type)
	}
}

func searchTool(cfg toolConfig, index index.Provider) (tool.Tool, error) {
	return search.New(index)
}
