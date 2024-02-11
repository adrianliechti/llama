package config

import (
	"errors"
	"os"
	"strings"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/chain/rag"
	"github.com/adrianliechti/llama/pkg/chain/react"
	"github.com/adrianliechti/llama/pkg/chain/refine"
	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
)

func (c *Config) registerChains(f *configFile) error {
	for id, cfg := range f.Chains {
		var err error

		var index index.Provider

		var prompt *prompt.Prompt

		var embedder provider.Embedder
		var completer provider.Completer

		classifiers := map[string]classifier.Provider{}

		if cfg.Model != "" {
			completer, err = c.Completer(cfg.Model)

			if err != nil {
				return err
			}
		}

		if cfg.Embedding != "" {
			embedder, err = c.Embedder(cfg.Embedding)

			if err != nil {
				return err
			}
		}

		if cfg.Index != "" {
			index, err = c.Index(cfg.Index)

			if err != nil {
				return err
			}
		}

		if cfg.Template != "" {
			prompt, err = parsePrompt(cfg.Template)

			if err != nil {
				return err
			}
		}

		for _, v := range cfg.Filters {
			if v.Classifier != "" {
				classifier, err := c.Classifier(v.Classifier)

				if err != nil {
					return err
				}

				classifiers[v.Classifier] = classifier
			}
		}

		chain, err := createChain(cfg, prompt, embedder, completer, index, classifiers)

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

func createChain(cfg chainConfig, prompt *prompt.Prompt, embedder provider.Embedder, completer provider.Completer, index index.Provider, classifiers map[string]classifier.Provider) (chain.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "rag":
		return ragChain(cfg, index, prompt, completer, classifiers)

	case "refine":
		return refineChain(cfg, index, prompt, completer, classifiers)

	case "fn", "react":
		return reactChain(cfg, prompt, completer)

	default:
		return nil, errors.New("invalid chain type: " + cfg.Type)
	}
}

func ragChain(cfg chainConfig, index index.Provider, prompt *prompt.Prompt, completer provider.Completer, classifiers map[string]classifier.Provider) (chain.Provider, error) {
	var options []rag.Option

	if index != nil {
		options = append(options, rag.WithIndex(index))
	}

	if prompt != nil {
		options = append(options, rag.WithPrompt(prompt))
	}

	if completer != nil {
		options = append(options, rag.WithCompleter(completer))
	}

	for k, v := range cfg.Filters {
		options = append(options, rag.WithFilter(k, classifiers[v.Classifier]))
	}

	if cfg.Limit != nil {
		options = append(options, rag.WithLimit(*cfg.Limit))
	}

	if cfg.Distance != nil {
		options = append(options, rag.WithDistance(*cfg.Distance))
	}

	return rag.New(options...)
}

func refineChain(cfg chainConfig, index index.Provider, prompt *prompt.Prompt, completer provider.Completer, classifiers map[string]classifier.Provider) (chain.Provider, error) {
	var options []refine.Option

	if index != nil {
		options = append(options, refine.WithIndex(index))
	}

	if prompt != nil {
		options = append(options, refine.WithPrompt(prompt))
	}
	if completer != nil {
		options = append(options, refine.WithCompleter(completer))
	}

	for k, v := range cfg.Filters {
		options = append(options, refine.WithFilter(k, classifiers[v.Classifier]))
	}

	if cfg.Limit != nil {
		options = append(options, refine.WithLimit(*cfg.Limit))
	}

	if cfg.Distance != nil {
		options = append(options, refine.WithDistance(*cfg.Distance))
	}

	return refine.New(options...)
}

func reactChain(cfg chainConfig, prompt *prompt.Prompt, completer provider.Completer) (chain.Provider, error) {
	var options []react.Option

	if prompt != nil {
		options = append(options, react.WithPrompt(prompt))
	}

	if completer != nil {
		options = append(options, react.WithCompleter(completer))
	}

	return react.New(options...)
}

func parsePrompt(val string) (*prompt.Prompt, error) {
	if val == "" {
		return nil, errors.New("empty prompt")
	}

	if data, err := os.ReadFile(val); err == nil {
		return prompt.New(string(data))
	}

	return prompt.New(val)
}
