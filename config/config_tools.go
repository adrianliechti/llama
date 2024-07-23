package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/adrianliechti/llama/pkg/tool/bing"
	"github.com/adrianliechti/llama/pkg/tool/custom"
	"github.com/adrianliechti/llama/pkg/tool/duckduckgo"
	"github.com/adrianliechti/llama/pkg/tool/tavily"
)

func (c *Config) RegisterTool(id string, val tool.Tool) {
	if c.tools == nil {
		c.tools = make(map[string]tool.Tool)
	}

	c.tools[id] = val
}

func (cfg *Config) registerTools(f *configFile) error {
	for id, t := range f.Tools {
		var err error

		tool, err := createTool(t)

		if err != nil {
			return err
		}

		cfg.RegisterTool(id, tool)
	}

	return nil
}

func createTool(cfg toolConfig) (tool.Tool, error) {
	switch strings.ToLower(cfg.Type) {
	case "bing":
		return bingTool(cfg)

	case "duckduckgo":
		return duckduckgoTool(cfg)

	case "tavily":
		return tavilyTool(cfg)

	case "custom":
		return customTool(cfg)

	default:
		return nil, errors.New("invalid tool type: " + cfg.Type)
	}
}

func bingTool(cfg toolConfig) (tool.Tool, error) {
	var options []bing.Option

	return bing.New(cfg.Token, options...)
}

func duckduckgoTool(cfg toolConfig) (tool.Tool, error) {
	var options []duckduckgo.Option

	return duckduckgo.New(options...)
}

func tavilyTool(cfg toolConfig) (tool.Tool, error) {
	var options []tavily.Option

	return tavily.New(cfg.Token, options...)
}

func customTool(cfg toolConfig) (tool.Tool, error) {
	var options []custom.Option

	return custom.New(cfg.URL, options...)
}
