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

func (c *Config) RegisterTool(id string, tool tool.Tool) {
	c.tools[id] = tool
}

type toolContext struct {
	Index     index.Provider
	Completer provider.Completer
}

func (c *Config) registerTools(f *configFile) error {
	for id, cfg := range f.Tools {
		var err error

		context := toolContext{}

		if cfg.Index != "" {
			if context.Index, err = c.Index(cfg.Index); err != nil {
				return err
			}
		}

		if cfg.Model != "" {
			if context.Completer, err = c.Completer(cfg.Model); err != nil {
				return err
			}
		}

		tool, err := createTool(cfg, context)

		if err != nil {
			return err
		}

		c.RegisterTool(id, tool)
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
