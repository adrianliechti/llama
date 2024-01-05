package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/classifier/llm"
	"github.com/adrianliechti/llama/pkg/provider"
)

func (c *Config) registerClassifiers(f *configFile) error {
	for id, cfg := range f.Classifiers {
		var err error

		var embedder provider.Embedder
		var completer provider.Completer

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

		classifier, err := createClassifier(cfg, embedder, completer)

		if err != nil {
			return err
		}

		c.classifiers[id] = classifier
	}

	return nil
}

func createClassifier(cfg classifierConfig, embedder provider.Embedder, completer provider.Completer) (classifier.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "llm":
		return llmClassifier(cfg, completer)

	default:
		return nil, errors.New("invalid index type: " + cfg.Type)
	}
}

func llmClassifier(cfg classifierConfig, completer provider.Completer) (classifier.Provider, error) {
	var options []llm.Option

	if cfg.Categories != nil {
		var categories []llm.Category

		for k, v := range cfg.Categories {
			categories = append(categories, llm.Category{
				Name:        k,
				Description: v,
			})
		}

		options = append(options, llm.WithCategories(categories...))
	}

	if completer != nil {
		options = append(options, llm.WithCompleter(completer))
	}

	return llm.New(options...)
}
