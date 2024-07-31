package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/azuretranslator"
	"github.com/adrianliechti/llama/pkg/provider/deepl"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (cfg *Config) RegisterTranslator(name, model string, p provider.Translator) {
	cfg.RegisterModel(model)

	if cfg.translator == nil {
		cfg.translator = make(map[string]provider.Translator)
	}

	translator, ok := p.(otel.ObservableTranslator)

	if !ok {
		translator = otel.NewTranslator(name, model, p)
	}

	cfg.translator[model] = translator
}

func (cfg *Config) Translator(model string) (provider.Translator, error) {
	if cfg.translator != nil {
		if t, ok := cfg.translator[model]; ok {
			return t, nil
		}
	}

	return nil, errors.New("translator not found: " + model)
}

func createTranslator(cfg providerConfig, model modelContext) (provider.Translator, error) {
	switch strings.ToLower(cfg.Type) {
	case "azuretranslator":
		return azureTranslator(cfg, model)

	case "deepl":
		return deeplTranslator(cfg, model)

	default:
		return nil, errors.New("invalid translator type: " + cfg.Type)
	}
}

func azureTranslator(cfg providerConfig, model modelContext) (provider.Translator, error) {
	var options []azuretranslator.Option

	if cfg.Token != "" {
		options = append(options, azuretranslator.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, azuretranslator.WithLanguage(model.ID))
	}

	return azuretranslator.NewTranslator(cfg.URL, options...)
}

func deeplTranslator(cfg providerConfig, model modelContext) (provider.Translator, error) {
	var options []deepl.Option

	if cfg.Token != "" {
		options = append(options, deepl.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, deepl.WithLanguage(model.ID))
	}

	return deepl.NewTranslator(cfg.URL, options...)
}
