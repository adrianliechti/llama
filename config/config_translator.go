package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/translator"
	"github.com/adrianliechti/llama/pkg/translator/azure"
	"github.com/adrianliechti/llama/pkg/translator/deepl"
)

func (cfg *Config) RegisterTranslator(model string, t translator.Provider) {
	cfg.RegisterModel(model)

	if cfg.translators == nil {
		cfg.translators = make(map[string]translator.Provider)
	}

	cfg.translators[model] = t
}

func (cfg *Config) registerTranslators(f *configFile) error {
	for _, p := range f.Providers {
		for id, m := range p.Models {
			translator, err := createTranslator(p, m.ID)

			if err != nil {
				return err
			}

			cfg.RegisterTranslator(id, translator)
		}
	}

	return nil
}

func createTranslator(cfg providerConfig, model string) (translator.Provider, error) {
	switch strings.ToLower(cfg.Type) {

	case "azure-translator":
		return azuretranslatorProvider(cfg, model)

	case "deepl":
		return deeplProvider(cfg, model)

	default:
		return nil, errors.New("invalid provider type: " + cfg.Type)
	}
}

func azuretranslatorProvider(cfg providerConfig, model string) (*azure.Client, error) {
	var options []azure.Option

	if cfg.Token != "" {
		options = append(options, azure.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, azure.WithLanguage(model))
	}

	return azure.New(cfg.URL, options...)
}

func deeplProvider(cfg providerConfig, model string) (*deepl.Client, error) {
	var options []deepl.Option

	if cfg.Token != "" {
		options = append(options, deepl.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, deepl.WithLanguage(model))
	}

	return deepl.New(cfg.URL, options...)
}
