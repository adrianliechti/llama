package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/classifier/llm"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
)

type classifierContext struct {
	Completer provider.Completer

	Template *prompt.Template
	Messages []provider.Message
}

func (c *Config) registerClassifiers(f *configFile) error {
	for id, cfg := range f.Classifiers {
		var err error

		context := classifierContext{}

		if cfg.Model != "" {
			if context.Completer, err = c.Completer(cfg.Model); err != nil {
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

		classifier, err := createClassifier(cfg, context)

		if err != nil {
			return err
		}

		c.classifiers[id] = classifier
	}

	return nil
}

func createClassifier(cfg classifierConfig, context classifierContext) (classifier.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "llm":
		return llmClassifier(cfg, context)

	default:
		return nil, errors.New("invalid index type: " + cfg.Type)
	}
}

func llmClassifier(cfg classifierConfig, context classifierContext) (classifier.Provider, error) {
	var options []llm.Option

	if context.Completer != nil {
		options = append(options, llm.WithCompleter(context.Completer))
	}

	if cfg.Classes != nil {
		var classes []classifier.Class

		for k, v := range cfg.Classes {
			classes = append(classes, classifier.Class{
				Name:        k,
				Description: v,
			})
		}

		options = append(options, llm.WithClasses(classes...))
	}

	return llm.New(options...)
}
