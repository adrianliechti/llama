package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/adrianliechti/llama/pkg/tool/bing"
	"github.com/adrianliechti/llama/pkg/tool/crawler"
	"github.com/adrianliechti/llama/pkg/tool/custom"
	"github.com/adrianliechti/llama/pkg/tool/draw"
	"github.com/adrianliechti/llama/pkg/tool/duckduckgo"
	"github.com/adrianliechti/llama/pkg/tool/retriever"
	"github.com/adrianliechti/llama/pkg/tool/searxng"
	"github.com/adrianliechti/llama/pkg/tool/tavily"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (c *Config) RegisterTool(alias string, p tool.Tool) {
	if c.tools == nil {
		c.tools = make(map[string]tool.Tool)
	}

	c.tools[alias] = p
}

func (cfg *Config) Tool(id string) (tool.Tool, error) {
	if cfg.tools != nil {
		if t, ok := cfg.tools[id]; ok {
			return t, nil
		}
	}

	return nil, errors.New("tool not found: " + id)
}

type toolConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Model string `yaml:"model"`

	Index     string `yaml:"index"`
	Extractor string `yaml:"extractor"`
}

type toolContext struct {
	Index     index.Provider
	Renderer  provider.Renderer
	Extractor extractor.Provider
}

func (cfg *Config) registerTools(f *configFile) error {
	for id, t := range f.Tools {
		var err error

		context := toolContext{}

		if i, err := cfg.Index(t.Index); err == nil {
			context.Index = i
		}

		if r, err := cfg.Renderer(t.Model); err == nil {
			context.Renderer = r
		}

		if e, err := cfg.Extractor(t.Extractor); err == nil {
			context.Extractor = e
		}

		tool, err := createTool(t, context)

		if err != nil {
			return err
		}

		if _, ok := tool.(otel.Tool); !ok {
			tool = otel.NewTool(t.Type, tool)
		}

		cfg.RegisterTool(id, tool)
	}

	return nil
}

func createTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	switch strings.ToLower(cfg.Type) {
	case "bing":
		return bingTool(cfg, context)

	case "crawler":
		return crawlerTool(cfg, context)

	case "draw":
		return drawTool(cfg, context)

	case "duckduckgo":
		return duckduckgoTool(cfg, context)

	case "retriever":
		return retrieverTool(cfg, context)

	case "searxng":
		return searxngTool(cfg, context)

	case "tavily":
		return tavilyTool(cfg, context)

	case "custom":
		return customTool(cfg, context)

	default:
		return nil, errors.New("invalid tool type: " + cfg.Type)
	}
}

func bingTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []bing.Option

	return bing.New(cfg.Token, options...)
}

func crawlerTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []crawler.Option

	return crawler.New(context.Extractor, options...)
}

func drawTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []draw.Option

	if context.Renderer != nil {
		options = append(options, draw.WithRenderer(context.Renderer))
	}

	return draw.New(options...)
}

func duckduckgoTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []duckduckgo.Option

	return duckduckgo.New(options...)
}

func retrieverTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []retriever.Option

	return retriever.New(context.Index, options...)
}

func searxngTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []searxng.Option

	return searxng.New(cfg.URL, options...)
}

func tavilyTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []tavily.Option

	return tavily.New(cfg.Token, options...)
}

func customTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []custom.Option

	return custom.New(cfg.URL, options...)
}
