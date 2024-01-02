package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
	"github.com/adrianliechti/llama/pkg/provider/sbert"
)

func (c *Config) registerProviders(f *configFile) error {
	for _, cfg := range f.Providers {
		p, err := createProvider(cfg)

		if err != nil {
			return err
		}

		for id, cfg := range cfg.Models {
			c.models[id] = Model{
				ID: id,

				model: cfg.ID,
			}

			c.providers[id] = p
		}
	}

	return nil
}

func createProvider(cfg providerConfig) (provider.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "openai":
		return openaiProvider(cfg)

	case "llama":
		return llamaProvider(cfg)

	case "ollama":
		return ollamaProvider(cfg)

	case "sbert":
		return sbertProvider(cfg)

	default:
		return nil, errors.New("invalid provider type: " + cfg.Type)
	}
}

func openaiProvider(cfg providerConfig) (provider.Provider, error) {
	var options []openai.Option

	if cfg.URL != "" {
		options = append(options, openai.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, openai.WithToken(cfg.Token))
	}

	return openai.New(options...)
}

func llamaProvider(cfg providerConfig) (provider.Provider, error) {
	var options []llama.Option

	var prompt string
	var template string

	for _, v := range cfg.Models {
		if prompt == "" {
			prompt = v.Prompt
		}

		if template == "" {
			template = v.Template
		}
	}

	if prompt != "" {
		options = append(options, llama.WithSystem(prompt))
	}

	switch strings.ToLower(template) {
	case "chatml":
		options = append(options, llama.WithTemplate(llama.TemplateChatML))

	case "llama":
		options = append(options, llama.WithTemplate(llama.TemplateLlama))

	case "llamaguard":
		options = append(options, llama.WithTemplate(llama.TemplateLlamaGuard))

	case "mistral":
		options = append(options, llama.WithTemplate(llama.TemplateMistral))

	default:
		return nil, errors.New("invalid prompt template: " + template)
	}

	return llama.New(cfg.URL, options...)
}

func ollamaProvider(cfg providerConfig) (provider.Provider, error) {
	var options []ollama.Option

	return ollama.New(cfg.URL, options...)
}

func sbertProvider(cfg providerConfig) (provider.Provider, error) {
	var options []sbert.Option

	if len(cfg.Models) > 1 {
		return nil, errors.New("multiple models not supported for sbert provider")
	}

	return sbert.New(cfg.URL, options...)
}
