package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/custom"
	"github.com/adrianliechti/llama/pkg/provider/langchain"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
	"github.com/adrianliechti/llama/pkg/provider/tei"
	"github.com/adrianliechti/llama/pkg/provider/tgi"
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
		return llamaProvider(cfg, model)

	case "whisper":
		return whisperProvider(cfg)

	case "ollama":
		return ollamaProvider(cfg, model)

	case "tei":
		return teiProvider(cfg)

	case "tgi":
		return tgiProvider(cfg)

	case "langchain":
		return langchainProvider(cfg, model)

	case "custom":
		return customProvider(cfg, model)

	default:
		return nil, errors.New("invalid provider type: " + cfg.Type)
	}
}

func openaiProvider(cfg providerConfig, model string) (*openai.Client, error) {
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

func llamaProvider(cfg providerConfig, model string) (*llama.Client, error) {
	var options []llama.Option

	if len(cfg.Models) != 1 {
		return nil, errors.New("llama supports exactly one model")
	}

	if model != "" {
		options = append(options, llama.WithModel(model))
	}

	return llama.New(cfg.URL, options...)
}

func ollamaProvider(cfg providerConfig, model string) (*ollama.Client, error) {
	var options []ollama.Option

	if model != "" {
		options = append(options, ollama.WithModel(model))
	}

	return ollama.New(cfg.URL, options...)
}

func langchainProvider(cfg providerConfig, model string) (*langchain.Client, error) {
	var options []langchain.Option

	// if model != "" {
	// 	options = append(options, langchain.WithModel(model))
	// }

	return langchain.New(cfg.URL, options...)
}

func customProvider(cfg providerConfig, model string) (*custom.Client, error) {
	var options []custom.Option

	return custom.New(cfg.URL, options...)
}

func teiProvider(cfg providerConfig) (*tei.Client, error) {
	var options []tei.Option

	if len(cfg.Models) > 1 {
		return nil, errors.New("multiple models not supported for tei provider")
	}

	return tei.New(cfg.URL, options...)
}

func tgiProvider(cfg providerConfig) (*tgi.Client, error) {
	var options []tgi.Option

	if len(cfg.Models) > 1 {
		return nil, errors.New("multiple models not supported for tgi provider")
	}

	return tgi.New(cfg.URL, options...)
}

func whisperProvider(cfg providerConfig) (*whisper.Client, error) {
	var options []whisper.Option

	if len(cfg.Models) > 1 {
		return nil, errors.New("multiple models not supported for tei provider")
	}

	return whisper.New(cfg.URL, options...)
}
