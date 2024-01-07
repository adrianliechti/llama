package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
	"github.com/adrianliechti/llama/pkg/provider/sbert"
	"github.com/adrianliechti/llama/pkg/provider/whisper"
)

func (c *Config) registerProviders(f *configFile) error {
	for _, p := range f.Providers {
		for id, m := range p.Models {
			r, err := createProvider(p, m.ID)

			if err != nil {
				return err
			}

			c.models[id] = provider.Model{
				ID: id,
			}

			if embedder, ok := r.(provider.Embedder); ok {
				c.embedder[id] = embedder
			}

			if completer, ok := r.(provider.Completer); ok {
				c.completer[id] = completer
			}

			if transcriber, ok := r.(provider.Transcriber); ok {
				c.transcriber[id] = transcriber
			}
		}
	}

	return nil
}

func createProvider(cfg providerConfig, model string) (any, error) {
	switch strings.ToLower(cfg.Type) {
	case "openai":
		return openaiProvider(cfg, model)

	case "llama":
		return llamaProvider(cfg)

	case "whisper":
		return whisperProvider(cfg)

	case "ollama":
		return ollamaProvider(cfg, model)

	case "sbert":
		return sbertProvider(cfg)

	default:
		return nil, errors.New("invalid provider type: " + cfg.Type)
	}
}

func openaiProvider(cfg providerConfig, model string) (*openai.Provider, error) {
	var options []openai.Option

	if cfg.URL != "" {
		options = append(options, openai.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, openai.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, openai.WithModel(model))
	}

	return openai.New(options...)
}

func ollamaProvider(cfg providerConfig, model string) (*ollama.Provider, error) {
	var options []ollama.Option

	if cfg.URL != "" {
		options = append(options, ollama.WithURL(cfg.URL))
	}

	if model != "" {
		options = append(options, ollama.WithModel(model))
	}

	return ollama.New(options...)
}

func llamaProvider(cfg providerConfig) (*llama.Provider, error) {
	var options []llama.Option

	var system string
	var template string

	for _, v := range cfg.Models {
		if system == "" {
			system = v.System
		}

		if template == "" {
			template = v.Template
		}
	}

	if system != "" {
		options = append(options, llama.WithSystem(system))
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

	case "simple":
		options = append(options, llama.WithTemplate(llama.TemplateSimple))

	default:
		return nil, errors.New("invalid prompt template: " + template)
	}

	return llama.New(cfg.URL, options...)
}

func whisperProvider(cfg providerConfig) (*whisper.Provider, error) {
	var options []whisper.Option

	return whisper.New(cfg.URL, options...)
}

func sbertProvider(cfg providerConfig) (*sbert.Provider, error) {
	var options []sbert.Option

	if len(cfg.Models) > 1 {
		return nil, errors.New("multiple models not supported for sbert provider")
	}

	return sbert.New(cfg.URL, options...)
}
