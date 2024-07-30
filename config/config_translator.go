package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/azuretranslator"
	"github.com/adrianliechti/llama/pkg/provider/deepl"
)

func (cfg *Config) RegisterTranslator(model string, t provider.Translator) {
	cfg.RegisterModel(model)

	if cfg.translator == nil {
		cfg.translator = make(map[string]provider.Translator)
	}

	cfg.translator[model] = t
}

func createTranslator(cfg providerConfig, model string) (provider.Translator, error) {
	switch strings.ToLower(cfg.Type) {

	case "azuretranslator":
		return azureTranslator(cfg, model)

	case "deepl":
		return deeplTranslator(cfg, model)

	default:
		return nil, errors.New("invalid translator type: " + cfg.Type)
	}
}

func azureTranslator(cfg providerConfig, model string) (provider.Translator, error) {
	var options []azuretranslator.Option

	if cfg.Token != "" {
		options = append(options, azuretranslator.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, azuretranslator.WithLanguage(model))
	}

	return azuretranslator.NewTranslator(cfg.URL, options...)
}

func deeplTranslator(cfg providerConfig, model string) (provider.Translator, error) {
	var options []deepl.Option

	if cfg.Token != "" {
		options = append(options, deepl.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, deepl.WithLanguage(model))
	}

	return deepl.NewTranslator(cfg.URL, options...)
}
