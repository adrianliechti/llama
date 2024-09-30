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
	"github.com/adrianliechti/llama/pkg/tool/speak"
	"github.com/adrianliechti/llama/pkg/tool/tavily"
	"github.com/adrianliechti/llama/pkg/tool/translate"
	"github.com/adrianliechti/llama/pkg/translator"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (c *Config) RegisterTool(id string, p tool.Tool) {
	if c.tools == nil {
		c.tools = make(map[string]tool.Tool)
	}

	c.tools[id] = p
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

	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Model    string `yaml:"model"`
	Provider string `yaml:"provider"`

	Index      string `yaml:"index"`
	Extractor  string `yaml:"extractor"`
	Translator string `yaml:"translator"`
}

type toolContext struct {
	Index      index.Provider
	Extractor  extractor.Provider
	Translator translator.Provider

	Renderer    provider.Renderer
	Synthesizer provider.Synthesizer
}

func (cfg *Config) registerTools(f *configFile) error {
	for id, t := range f.Tools {
		var err error

		context := toolContext{}

		if p, err := cfg.Index(t.Index); err == nil {
			context.Index = p
		}

		if p, err := cfg.Extractor(t.Extractor); err == nil {
			context.Extractor = p
		}

		if p, err := cfg.Translator(t.Translator); err == nil {
			context.Translator = p
		}

		if p, err := cfg.Index(t.Provider); err == nil {
			context.Index = p
		}

		if p, err := cfg.Extractor(t.Provider); err == nil {
			context.Extractor = p
		}

		if p, err := cfg.Translator(t.Provider); err == nil {
			context.Translator = p
		}

		if p, err := cfg.Renderer(t.Model); err == nil {
			context.Renderer = p
		}

		if p, err := cfg.Synthesizer(t.Model); err == nil {
			context.Synthesizer = p
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

	case "speak":
		return speakTool(cfg, context)

	case "tavily":
		return tavilyTool(cfg, context)

	case "translate":
		return translateTool(cfg, context)

	case "custom":
		return customTool(cfg, context)

	default:
		return nil, errors.New("invalid tool type: " + cfg.Type)
	}
}

func bingTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []bing.Option

	if cfg.Name != "" {
		options = append(options, bing.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, bing.WithDescription(cfg.Description))
	}

	return bing.New(cfg.Token, options...)
}

func crawlerTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []crawler.Option

	if cfg.Name != "" {
		options = append(options, crawler.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, crawler.WithDescription(cfg.Description))
	}

	return crawler.New(context.Extractor, options...)
}

func drawTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []draw.Option

	if cfg.Name != "" {
		options = append(options, draw.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, draw.WithDescription(cfg.Description))
	}

	if context.Renderer != nil {
		options = append(options, draw.WithRenderer(context.Renderer))
	}

	return draw.New(options...)
}

func duckduckgoTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []duckduckgo.Option

	if cfg.Name != "" {
		options = append(options, duckduckgo.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, duckduckgo.WithDescription(cfg.Description))
	}

	return duckduckgo.New(options...)
}

func retrieverTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []retriever.Option

	if cfg.Name != "" {
		options = append(options, retriever.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, retriever.WithDescription(cfg.Description))
	}

	return retriever.New(context.Index, options...)
}

func searxngTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []searxng.Option

	if cfg.Name != "" {
		options = append(options, searxng.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, searxng.WithDescription(cfg.Description))
	}

	return searxng.New(cfg.URL, options...)
}

func speakTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []speak.Option

	if cfg.Name != "" {
		options = append(options, speak.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, speak.WithDescription(cfg.Description))
	}

	if context.Synthesizer != nil {
		options = append(options, speak.WithSynthesizer(context.Synthesizer))
	}

	return speak.New(options...)
}

func tavilyTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []tavily.Option

	if cfg.Name != "" {
		options = append(options, tavily.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, tavily.WithDescription(cfg.Description))
	}

	return tavily.New(cfg.Token, options...)
}

func translateTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []translate.Option

	if cfg.Name != "" {
		options = append(options, translate.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, translate.WithDescription(cfg.Description))
	}

	return translate.New(context.Translator, options...)
}

func customTool(cfg toolConfig, context toolContext) (tool.Tool, error) {
	var options []custom.Option

	if cfg.Name != "" {
		options = append(options, custom.WithName(cfg.Name))
	}

	if cfg.Description != "" {
		options = append(options, custom.WithDescription(cfg.Description))
	}

	return custom.New(cfg.URL, options...)
}
