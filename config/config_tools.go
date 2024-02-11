package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/adrianliechti/llama/pkg/tool/tavily"
)

func (c *Config) registerTools(f *configFile) error {
	for id, cfg := range f.Tools {
		t, err := createTool(cfg)

		if err != nil {
			return err
		}

		c.tools[id] = t
	}

	return nil
}

func createTool(cfg toolConfig) (tool.Tool, error) {
	switch strings.ToLower(cfg.Type) {
	case "tavily":
		return tavilyTool(cfg)

	default:
		return nil, errors.New("invalid tool type: " + cfg.Type)
	}
}

func tavilyTool(cfg toolConfig) (tool.Tool, error) {
	return tavily.New(cfg.Token)
}
