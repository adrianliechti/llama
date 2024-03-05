package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/anthropic"
	"github.com/adrianliechti/llama/pkg/provider/azuretranslator"
	"github.com/adrianliechti/llama/pkg/provider/custom"
	"github.com/adrianliechti/llama/pkg/provider/deepl"
	"github.com/adrianliechti/llama/pkg/provider/huggingface"
	"github.com/adrianliechti/llama/pkg/provider/langchain"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
	"github.com/adrianliechti/llama/pkg/provider/whisper"
)

func (cfg *Config) RegisterEmbedder(model string, e provider.Embedder) {
	cfg.RegisterModel(model)

	if cfg.embedder == nil {
		cfg.embedder = make(map[string]provider.Embedder)
	}

	cfg.embedder[model] = e
}

func (cfg *Config) RegisterCompleter(model string, c provider.Completer) {
	cfg.RegisterModel(model)

	if cfg.completer == nil {
		cfg.completer = make(map[string]provider.Completer)
	}

	cfg.completer[model] = c
}

func (cfg *Config) RegisterTranslator(model string, t provider.Translator) {
	cfg.RegisterModel(model)

	if cfg.translator == nil {
		cfg.translator = make(map[string]provider.Translator)
	}

	cfg.translator[model] = t
}

func (cfg *Config) RegisterTranscriber(model string, t provider.Transcriber) {
	cfg.RegisterModel(model)

	if cfg.transcriber == nil {
		cfg.transcriber = make(map[string]provider.Transcriber)
	}

	cfg.transcriber[model] = t
}

func (cfg *Config) registerProviders(f *configFile) error {
	for _, p := range f.Providers {
		for id, m := range p.Models {
			r, err := createProvider(p, m.ID)

			if err != nil {
				return err
			}

			if embedder, ok := r.(provider.Embedder); ok {
				cfg.RegisterEmbedder(id, embedder)
			}

			if completer, ok := r.(provider.Completer); ok {
				cfg.RegisterCompleter(id, completer)
			}

			if translator, ok := r.(provider.Translator); ok {
				cfg.RegisterTranslator(id, translator)
			}

			if transcriber, ok := r.(provider.Transcriber); ok {
				cfg.RegisterTranscriber(id, transcriber)
			}
		}
	}

	return nil
}

func createProvider(cfg providerConfig, model string) (any, error) {
	switch strings.ToLower(cfg.Type) {
	case "openai":
		return openaiProvider(cfg, model)

	case "anthropic":
		return anthropicProvider(cfg, model)

	case "llama":
		return llamaProvider(cfg, model)

	case "whisper":
		return whisperProvider(cfg)

	case "ollama":
		return ollamaProvider(cfg, model)

	case "huggingface":
		return huggingfaceProvider(cfg, model)

	case "langchain":
		return langchainProvider(cfg, model)

	case "deepl":
		return deeplProvider(cfg, model)

	case "azure-translator":
		return azuretranslatorProvider(cfg, model)

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

func anthropicProvider(cfg providerConfig, model string) (*anthropic.Client, error) {
	var options []anthropic.Option

	// if cfg.URL != "" {
	// 	options = append(options, openai.WithURL(cfg.URL))
	// }

	if cfg.Token != "" {
		options = append(options, anthropic.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, anthropic.WithModel(model))
	}

	return anthropic.New(options...)
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

func huggingfaceProvider(cfg providerConfig, model string) (*huggingface.Client, error) {
	var options []huggingface.Option

	if cfg.Token != "" {
		options = append(options, huggingface.WithToken(cfg.Token))
	}

	if len(cfg.Models) > 1 {
		return nil, errors.New("multiple models not supported for hugging face provider")
	}

	return huggingface.New(cfg.URL, options...)
}

func whisperProvider(cfg providerConfig) (*whisper.Client, error) {
	var options []whisper.Option

	if len(cfg.Models) > 1 {
		return nil, errors.New("multiple models not supported for tei provider")
	}

	return whisper.New(cfg.URL, options...)
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

func azuretranslatorProvider(cfg providerConfig, model string) (*azuretranslator.Client, error) {
	var options []azuretranslator.Option

	if cfg.Token != "" {
		options = append(options, azuretranslator.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, azuretranslator.WithLanguage(model))
	}

	return azuretranslator.New(cfg.URL, options...)
}
