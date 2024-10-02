package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/translator"
	"github.com/adrianliechti/llama/pkg/translator/azure"
	"github.com/adrianliechti/llama/pkg/translator/deepl"
	"golang.org/x/time/rate"
)

func (cfg *Config) RegisterTranslator(id string, p translator.Provider) {
	if cfg.translator == nil {
		cfg.translator = make(map[string]translator.Provider)
	}

	cfg.translator[id] = p
}

func (cfg *Config) Translator(id string) (translator.Provider, error) {
	if cfg.translator != nil {
		if t, ok := cfg.translator[id]; ok {
			return t, nil
		}
	}

	return nil, errors.New("translator not found: " + id)
}

type translatorConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Limit *int `yaml:"limit"`
}

type translatorContext struct {
	Limiter *rate.Limiter
}

func (cfg *Config) RegisterTranslators(f *configFile) error {
	var translators []translator.Provider

	for id, t := range f.Translators {
		context := translatorContext{}

		limit := t.Limit

		if limit == nil {
			limit = t.Limit
		}

		if limit != nil {
			context.Limiter = rate.NewLimiter(rate.Limit(*limit), *limit)
		}

		translator, err := createTranslator(t, context)

		if err != nil {
			return err
		}

		translators = append(translators, translator)

		cfg.RegisterTranslator(id, translator)
	}

	return nil
}

func createTranslator(cfg translatorConfig, context translatorContext) (translator.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "azure":
		return azureTranslator(cfg, context)

	case "deepl":
		return deeplTranslator(cfg, context)

	default:
		return nil, errors.New("invalid translator type: " + cfg.Type)
	}
}

func azureTranslator(cfg translatorConfig, context translatorContext) (translator.Provider, error) {
	var options []azure.Option

	if cfg.Token != "" {
		options = append(options, azure.WithToken(cfg.Token))
	}

	return azure.NewTranslator(cfg.URL, options...)
}

func deeplTranslator(cfg translatorConfig, context translatorContext) (translator.Provider, error) {
	var options []deepl.Option

	if cfg.Token != "" {
		options = append(options, deepl.WithToken(cfg.Token))
	}

	return deepl.NewTranslator(cfg.URL, options...)
}
