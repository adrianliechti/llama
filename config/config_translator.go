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

	if _, ok := cfg.translator[""]; !ok {
		cfg.translator[""] = p
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

func (cfg *Config) registerTranslators(f *configFile) error {
	var configs map[string]translatorConfig

	if err := f.Translators.Decode(&configs); err != nil {
		return err
	}

	for _, node := range f.Translators.Content {
		id := node.Value

		config, ok := configs[node.Value]

		if !ok {
			continue
		}

		context := translatorContext{
			Limiter: createLimiter(config.Limit),
		}

		translator, err := createTranslator(config, context)

		if err != nil {
			return err
		}

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
