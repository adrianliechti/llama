package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/translator"
	"github.com/adrianliechti/llama/pkg/translator/azure"
	"github.com/adrianliechti/llama/pkg/translator/deepl"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (cfg *Config) RegisterTranslator(name, model string, p translator.Provider) {
	if cfg.translator == nil {
		cfg.translator = make(map[string]translator.Provider)
	}

	translator, ok := p.(otel.ObservableTranslator)

	if !ok {
		translator = otel.NewTranslator(name, model, p)
	}

	cfg.translator[model] = translator
}

func (cfg *Config) Translator(model string) (translator.Provider, error) {
	if cfg.translator != nil {
		if t, ok := cfg.translator[model]; ok {
			return t, nil
		}
	}

	return nil, errors.New("translator not found: " + model)
}

func createTranslator(cfg providerConfig, model modelContext) (translator.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "azure":
		return azureTranslator(cfg, model)

	case "deepl":
		return deeplTranslator(cfg, model)

	default:
		return nil, errors.New("invalid translator type: " + cfg.Type)
	}
}

func azureTranslator(cfg providerConfig, model modelContext) (translator.Provider, error) {
	var options []azure.Option

	if model.ID != "" {
		options = append(options, azure.WithLanguage(model.ID))
	}

	return azure.NewTranslator(cfg.URL, cfg.Token, options...)
}

func deeplTranslator(cfg providerConfig, model modelContext) (translator.Provider, error) {
	var options []deepl.Option

	if cfg.Token != "" {
		options = append(options, deepl.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, deepl.WithLanguage(model.ID))
	}

	return deepl.NewTranslator(cfg.URL, options...)
}
