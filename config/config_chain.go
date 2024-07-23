package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/chain/agent"
	"github.com/adrianliechti/llama/pkg/chain/assistant"
	"github.com/adrianliechti/llama/pkg/chain/rag"

	"github.com/adrianliechti/llama/pkg/to"
	"github.com/adrianliechti/llama/pkg/tool"
)

func (cfg *Config) RegisterChain(model string, c chain.Provider) {
	cfg.RegisterModel(model)

	if cfg.chains == nil {
		cfg.chains = make(map[string]chain.Provider)
	}

	cfg.chains[model] = c
}

type chainContext struct {
	Index index.Provider

	Embedder  provider.Embedder
	Completer provider.Completer

	Template *prompt.Template
	Messages []provider.Message

	Tools map[string]tool.Tool

	Classifiers map[string]classifier.Provider
}

func (cfg *Config) registerChains(f *configFile) error {
	for id, c := range f.Chains {
		var err error

		context := chainContext{
			Tools:       make(map[string]tool.Tool),
			Messages:    make([]provider.Message, 0),
			Classifiers: make(map[string]classifier.Provider),
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

			context.Tools[tool.Name()] = tool
		}

		for _, v := range c.Filters {
			if v.Classifier != "" {
				classifier, err := cfg.Classifier(v.Classifier)

				if err != nil {
					return err
				}

				context.Classifiers[v.Classifier] = classifier
			}
		}

		chain, err := createChain(c, context)

		if err != nil {
			return err
		}

		cfg.RegisterChain(id, chain)
	}

	return nil
}

func createChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "assistant":
		return assistantChain(cfg, context)

	case "rag":
		return ragChain(cfg, context)

	case "agent":
		return agentChain(cfg, context)

	default:
		return nil, errors.New("invalid chain type: " + cfg.Type)
	}
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

	if cfg.Limit != nil {
		options = append(options, rag.WithLimit(*cfg.Limit))
	}

	if cfg.Distance != nil {
		options = append(options, rag.WithDistance(*cfg.Distance))
	}

	if cfg.Temperature != nil {
		options = append(options, rag.WithTemperature(*cfg.Temperature))
	}

	for k, v := range cfg.Filters {
		options = append(options, rag.WithFilter(k, context.Classifiers[v.Classifier]))
	}

	return rag.New(options...)
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
