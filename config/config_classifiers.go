package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/classifier/llm"
	"github.com/adrianliechti/llama/pkg/prompt"
	"github.com/adrianliechti/llama/pkg/provider"
)

func (cfg *Config) RegisterClassifier(model string, c classifier.Provider) {
	if cfg.classifiers == nil {
		cfg.classifiers = make(map[string]classifier.Provider)
	}

	cfg.classifiers[model] = c
}

type classifierContext struct {
	Completer provider.Completer

	Template *prompt.Template
	Messages []provider.Message
}

func (cfg *Config) registerClassifiers(f *configFile) error {
	for id, c := range f.Classifiers {
		var err error

		context := classifierContext{}

		if c.Model != "" {
			if context.Completer, err = cfg.Completer(c.Model); err != nil {
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

		classifier, err := createClassifier(c, context)

		if err != nil {
			return err
		}

		cfg.RegisterClassifier(id, classifier)
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
