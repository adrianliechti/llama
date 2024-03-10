package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/adrianliechti/llama/pkg/tool/custom"
	"github.com/adrianliechti/llama/pkg/tool/search"
)

func (c *Config) RegisterTool(id string, val tool.Tool) {
	if c.tools == nil {
		c.tools = make(map[string]tool.Tool)
	}

	c.tools[id] = val
}

type toolContext struct {
	Index     index.Provider
	Completer provider.Completer
}

func (cfg *Config) registerTools(f *configFile) error {
	for id, t := range f.Tools {
		var err error

		context := toolContext{}

		if t.Index != "" {
			if context.Index, err = cfg.Index(t.Index); err != nil {
				return err
			}
		}

		if t.Model != "" {
			if context.Completer, err = cfg.Completer(t.Model); err != nil {
				return err
			}
		}

		tool, err := createTool(t, context)

		if err != nil {
			return err
		}

		cfg.RegisterTool(id, tool)
	}

	return nil
}

func createTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	switch strings.ToLower(cfg.Type) {
	case "search":
		return searchTool(cfg, context)

	case "custom":
		return customTool(cfg, context)

	default:
		return nil, errors.New("invalid tool type: " + cfg.Type)
	}
}

func searchTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	return search.New(context.Index)
}

func customTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []custom.Option

	return custom.New(cfg.URL, options...)
}
