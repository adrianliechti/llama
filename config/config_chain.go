package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/limiter"
	"github.com/adrianliechti/llama/pkg/otel"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
	"golang.org/x/time/rate"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/chain/agent"
	"github.com/adrianliechti/llama/pkg/chain/assistant"
	"github.com/adrianliechti/llama/pkg/chain/rag"
	"github.com/adrianliechti/llama/pkg/chain/reasoning"

	"github.com/adrianliechti/llama/pkg/to"
	"github.com/adrianliechti/llama/pkg/tool"
)

func (cfg *Config) RegisterChain(model string, p chain.Provider) {
	cfg.RegisterModel(model)

	if cfg.chains == nil {
		cfg.chains = make(map[string]chain.Provider)
	}

	cfg.chains[model] = p
}

type chainConfig struct {
	Type string `yaml:"type"`

	Model     string `yaml:"model"`
	Index     string `yaml:"index"`
	Embedding string `yaml:"embedding"`

	Template string    `yaml:"template"`
	Messages []message `yaml:"messages"`

	Tools []string `yaml:"tools"`

	Limit       *int     `yaml:"limit"`
	Temperature *float32 `yaml:"temperature"`
}

type chainContext struct {
	Index index.Provider

	Embedder  provider.Embedder
	Completer provider.Completer

	Template *prompt.Template
	Messages []provider.Message

	Tools map[string]tool.Tool

	Limiter *rate.Limiter
}

func (cfg *Config) registerChains(f *configFile) error {
	for id, c := range f.Chains {
		var err error

		context := chainContext{
			Tools:    make(map[string]tool.Tool),
			Messages: make([]provider.Message, 0),
		}

		limit := c.Limit

		if limit == nil {
			limit = c.Limit
		}

		if limit != nil {
			context.Limiter = rate.NewLimiter(rate.Limit(*limit), *limit)
		}

		if c.Index != "" {
			if context.Index, err = cfg.Index(c.Index); err != nil {
				return err
			}
		}

		if c.Model != "" {
			if context.Completer, err = cfg.Completer(c.Model); err != nil {
				return err
			}
		}

		if c.Embedding != "" {
			if context.Embedder, err = cfg.Embedder(c.Embedding); err != nil {
				return err
			}
		}

		if c.Template != "" {
			if context.Template, err = parseTemplate(c.Template); err != nil {
				return err
			}
		}

		if c.Messages != nil {
			if context.Messages, err = parseMessages(c.Messages); err != nil {
				return err
			}
		}

		for _, t := range c.Tools {
			tool, err := cfg.Tool(t)

			if err != nil {
				return err
			}

			context.Tools[t] = tool
		}

		chain, err := createChain(c, context)

		if err != nil {
			return err
		}

		if _, ok := chain.(limiter.Chain); !ok {
			chain = limiter.NewChain(context.Limiter, chain)
		}

		if _, ok := chain.(otel.Chain); !ok {
			chain = otel.NewChain(c.Type, id, chain)
		}

		cfg.RegisterChain(id, chain)
	}

	return nil
}

func createChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "agent":
		return agentChain(cfg, context)

	case "assistant":
		return assistantChain(cfg, context)

	case "rag":
		return ragChain(cfg, context)

	case "reasoning":
		return reasoningChain(cfg, context)

	default:
		return nil, errors.New("invalid chain type: " + cfg.Type)
	}
}

func agentChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	var options []agent.Option

	if context.Completer != nil {
		options = append(options, agent.WithCompleter(context.Completer))
	}

	if context.Tools != nil {
		options = append(options, agent.WithTools(to.Values(context.Tools)...))
	}

	if cfg.Temperature != nil {
		options = append(options, agent.WithTemperature(*cfg.Temperature))
	}

	return agent.New(options...)
}

func assistantChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	var options []assistant.Option

	if context.Completer != nil {
		options = append(options, assistant.WithCompleter(context.Completer))
	}

	if context.Template != nil {
		options = append(options, assistant.WithTemplate(context.Template))
	}

	if context.Messages != nil {
		options = append(options, assistant.WithMessages(context.Messages...))
	}

	if cfg.Temperature != nil {
		options = append(options, assistant.WithTemperature(*cfg.Temperature))
	}

	return assistant.New(options...)
}

func ragChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	var options []rag.Option

	if context.Completer != nil {
		options = append(options, rag.WithCompleter(context.Completer))
	}

	if context.Template != nil {
		options = append(options, rag.WithTemplate(context.Template))
	}

	if context.Messages != nil {
		options = append(options, rag.WithMessages(context.Messages...))
	}

	if context.Index != nil {
		options = append(options, rag.WithIndex(context.Index))
	}

	if cfg.Temperature != nil {
		options = append(options, rag.WithTemperature(*cfg.Temperature))
	}

	return rag.New(options...)
}

func reasoningChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	var options []reasoning.Option

	if context.Completer != nil {
		options = append(options, reasoning.WithCompleter(context.Completer))
	}

	if cfg.Temperature != nil {
		options = append(options, reasoning.WithTemperature(*cfg.Temperature))
	}

	return reasoning.New(options...)
}
