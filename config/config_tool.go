package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/adrianliechti/llama/pkg/tool"
	"github.com/adrianliechti/llama/pkg/tool/crawler"
	"github.com/adrianliechti/llama/pkg/tool/custom"
	"github.com/adrianliechti/llama/pkg/tool/draw"
	"github.com/adrianliechti/llama/pkg/tool/memory"
	"github.com/adrianliechti/llama/pkg/tool/retriever"
	"github.com/adrianliechti/llama/pkg/tool/search"
	"github.com/adrianliechti/llama/pkg/tool/speak"
	"github.com/adrianliechti/llama/pkg/tool/translate"
	"github.com/adrianliechti/llama/pkg/translator"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/index/bing"
	"github.com/adrianliechti/llama/pkg/index/duckduckgo"
	"github.com/adrianliechti/llama/pkg/index/searxng"
	"github.com/adrianliechti/llama/pkg/index/tavily"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (c *Config) RegisterTool(id string, p tool.Provider) {
	if c.tools == nil {
		c.tools = make(map[string]tool.Provider)
	}

	c.tools[id] = p
}

func (cfg *Config) Tool(id string) (tool.Provider, error) {
	if cfg.tools != nil {
		if p, ok := cfg.tools[id]; ok {
			return p, nil
		}
	}

	return nil, errors.New("tool not found: " + id)
}

type toolConfig struct {
	Type string `yaml:"type"`

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
	var configs map[string]toolConfig

	if err := f.Tools.Decode(&configs); err != nil {
		return err
	}

	for _, node := range f.Tools.Content {
		id := node.Value

		config, ok := configs[node.Value]

		if !ok {
			continue
		}

		context := toolContext{}

		if p, err := cfg.Index(config.Index); err == nil {
			context.Index = p
		}

		if p, err := cfg.Extractor(config.Extractor); err == nil {
			context.Extractor = p
		}

		if p, err := cfg.Translator(config.Translator); err == nil {
			context.Translator = p
		}

		if p, err := cfg.Index(config.Provider); err == nil {
			context.Index = p
		}

		if p, err := cfg.Extractor(config.Provider); err == nil {
			context.Extractor = p
		}

		if p, err := cfg.Translator(config.Provider); err == nil {
			context.Translator = p
		}

		if p, err := cfg.Renderer(config.Model); err == nil {
			context.Renderer = p
		}

		if p, err := cfg.Synthesizer(config.Model); err == nil {
			context.Synthesizer = p
		}

		tool, err := createTool(config, context)

		if err != nil {
			return err
		}

		if _, ok := tool.(otel.Tool); !ok {
			tool = otel.NewTool(config.Type, tool)
		}

		cfg.RegisterTool(id, tool)
	}

	return nil
}

func createTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	switch strings.ToLower(cfg.Type) {

	case "crawler":
		return crawlerTool(cfg, context)

	case "draw":
		return drawTool(cfg, context)

	case "memory":
		return memoryTool(cfg, context)

	case "retriever":
		return retrieverTool(cfg, context)

	case "search":
		return searchTool(cfg, context)

	case "speak":
		return speakTool(cfg, context)

	case "translate":
		return translateTool(cfg, context)

	case "custom":
		return customTool(cfg, context)

	case "bing":
		return bingTool(cfg, context)

	case "duckduckgo":
		return duckduckgoTool(cfg, context)

	case "searxng":
		return searxngTool(cfg, context)

	case "tavily":
		return tavilyTool(cfg, context)

	default:
		return nil, errors.New("invalid tool type: " + cfg.Type)
	}
}

func crawlerTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	var options []crawler.Option

	return crawler.New(context.Extractor, options...)
}

func drawTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	var options []draw.Option

	return draw.New(context.Renderer, options...)
}

func memoryTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	var options []memory.Option

	return memory.New(options...)
}

func retrieverTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	var options []retriever.Option

	return retriever.New(context.Index, options...)
}

func searchTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	var options []search.Option

	return search.New(context.Index, options...)
}

func speakTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	var options []speak.Option

	return speak.New(context.Synthesizer, options...)
}

func translateTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	var options []translate.Option

	return translate.New(context.Translator, options...)
}

func customTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	var options []custom.Option

	return custom.New(cfg.URL, options...)
}

func bingTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	index, err := bing.New(cfg.Token)

	if err != nil {
		return nil, err
	}

	context.Index = index

	return searchTool(cfg, context)
}

func duckduckgoTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	index, err := duckduckgo.New()

	if err != nil {
		return nil, err
	}

	context.Index = index

	return searchTool(cfg, context)
}

func searxngTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	index, err := searxng.New(cfg.Token)

	if err != nil {
		return nil, err
	}

	context.Index = index

	return searchTool(cfg, context)
}

func tavilyTool(cfg toolConfig, context toolContext) (tool.Provider, error) {
	index, err := tavily.New(cfg.Token)

	if err != nil {
		return nil, err
	}

	context.Index = index

	return searchTool(cfg, context)
}
