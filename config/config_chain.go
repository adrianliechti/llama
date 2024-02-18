package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/chain/assistant"
	"github.com/adrianliechti/llama/pkg/chain/rag"
	"github.com/adrianliechti/llama/pkg/chain/react"
	"github.com/adrianliechti/llama/pkg/chain/refine"
	"github.com/adrianliechti/llama/pkg/chain/toolbox"
	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/to"
	"github.com/adrianliechti/llama/pkg/tool"
)

type chainContext struct {
	Index index.Provider

	Embedder  provider.Embedder
	Completer provider.Completer

	Template *prompt.Template
	Messages []provider.Message

	Tools map[string]tool.Tool

	Classifiers map[string]classifier.Provider
}

func (c *Config) registerChains(f *configFile) error {
	for id, cfg := range f.Chains {
		var err error

		context := chainContext{
			Tools:       make(map[string]tool.Tool),
			Messages:    make([]provider.Message, 0),
			Classifiers: make(map[string]classifier.Provider),
		}

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

		if cfg.Embedding != "" {
			if context.Embedder, err = c.Embedder(cfg.Embedding); err != nil {
				return err
			}
		}

		if cfg.Template != "" {
			if context.Template, err = parseTemplate(cfg.Template); err != nil {
				return err
			}
		}

		if cfg.Messages != nil {
			if context.Messages, err = parseMessages(cfg.Messages); err != nil {
				return err
			}
		}

		for _, t := range cfg.Tools {
			tool, err := c.Tool(t)

			if err != nil {
				return err
			}

			context.Tools[tool.Name()] = tool
		}

		for _, v := range cfg.Filters {
			if v.Classifier != "" {
				classifier, err := c.Classifier(v.Classifier)

				if err != nil {
					return err
				}

				context.Classifiers[v.Classifier] = classifier
			}
		}

		chain, err := createChain(cfg, context)

		if err != nil {
			return err
		}

		c.models[id] = provider.Model{
			ID: id,
		}

		c.chains[id] = chain
	}

	return nil
}

func createChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "assistant":
		return assistantChain(cfg, context)

	case "rag":
		return ragChain(cfg, context)

	case "refine":
		return refineChain(cfg, context)

	case "react":
		return reactChain(cfg, context)

	case "toolbox":
		return toolboxChain(cfg, context)

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

	for k, v := range cfg.Filters {
		options = append(options, rag.WithFilter(k, context.Classifiers[v.Classifier]))
	}

	return rag.New(options...)
}

func refineChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	var options []refine.Option

	if context.Completer != nil {
		options = append(options, refine.WithCompleter(context.Completer))
	}

	if context.Template != nil {
		options = append(options, refine.WithTemplate(context.Template))
	}

	if context.Messages != nil {
		options = append(options, refine.WithMessages(context.Messages...))
	}

	if context.Index != nil {
		options = append(options, refine.WithIndex(context.Index))
	}

	if cfg.Limit != nil {
		options = append(options, refine.WithLimit(*cfg.Limit))
	}

	if cfg.Distance != nil {
		options = append(options, refine.WithDistance(*cfg.Distance))
	}

	for k, v := range cfg.Filters {
		options = append(options, refine.WithFilter(k, context.Classifiers[v.Classifier]))
	}

	return refine.New(options...)
}

func reactChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	var options []react.Option

	if context.Completer != nil {
		options = append(options, react.WithCompleter(context.Completer))
	}

	if context.Template != nil {
		options = append(options, react.WithTemplate(context.Template))
	}

	if context.Messages != nil {
		options = append(options, react.WithMessages(context.Messages...))
	}

	return react.New(options...)
}

func toolboxChain(cfg chainConfig, context chainContext) (chain.Provider, error) {
	var options []toolbox.Option

	if context.Completer != nil {
		options = append(options, toolbox.WithCompleter(context.Completer))
	}

	if context.Tools != nil {
		options = append(options, toolbox.WithTools(to.Values(context.Tools)...))
	}

	return toolbox.New(options...)
}
