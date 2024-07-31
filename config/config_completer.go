package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/anthropic"
	"github.com/adrianliechti/llama/pkg/provider/cohere"
	"github.com/adrianliechti/llama/pkg/provider/custom"
	"github.com/adrianliechti/llama/pkg/provider/groq"
	"github.com/adrianliechti/llama/pkg/provider/huggingface"
	"github.com/adrianliechti/llama/pkg/provider/langchain"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/mistral"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
	"github.com/adrianliechti/llama/pkg/provider/openai"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (cfg *Config) RegisterCompleter(name, model string, p provider.Completer) {
	cfg.RegisterModel(model)

	if cfg.completer == nil {
		cfg.completer = make(map[string]provider.Completer)
	}

	completer, ok := p.(otel.ObservableCompleter)

	if !ok {
		completer = otel.NewCompleter(name, model, p)
	}

	cfg.completer[model] = completer
}

func (cfg *Config) Completer(model string) (provider.Completer, error) {
	if cfg.completer != nil {
		if c, ok := cfg.completer[model]; ok {
			return c, nil
		}
	}

	if cfg.chains != nil {
		if c, ok := cfg.chains[model]; ok {
			return c, nil
		}
	}

	return nil, errors.New("completer not found: " + model)
}

func createCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	switch strings.ToLower(cfg.Type) {
	case "anthropic":
		return anthropicCompleter(cfg, model)

	case "cohere":
		return cohereCompleter(cfg, model)

	case "groq":
		return groqCompleter(cfg, model)

	case "huggingface":
		return huggingfaceCompleter(cfg, model)

	case "langchain":
		return langchainCompleter(cfg, model)

	case "llama":
		return llamaCompleter(cfg, model)

	case "mistral":
		return mistralCompleter(cfg, model)

	case "ollama":
		return ollamaCompleter(cfg, model)

	case "openai":
		return openaiCompleter(cfg, model)

	case "custom":
		return customCompleter(cfg, model)

	default:
		return nil, errors.New("invalid completer type: " + cfg.Type)
	}
}

func anthropicCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []anthropic.Option

	if cfg.URL != "" {
		options = append(options, anthropic.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, anthropic.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, anthropic.WithModel(model.ID))
	}

	return anthropic.NewCompleter(options...)
}

func cohereCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []cohere.Option

	if cfg.Token != "" {
		options = append(options, cohere.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, cohere.WithModel(model.ID))
	}

	return cohere.NewCompleter(options...)
}

func groqCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []groq.Option

	if cfg.Token != "" {
		options = append(options, groq.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, groq.WithModel(model.ID))
	}

	return groq.NewCompleter(options...)
}

func huggingfaceCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []huggingface.Option

	if cfg.Token != "" {
		options = append(options, huggingface.WithToken(cfg.Token))
	}

	return huggingface.NewCompleter(cfg.URL, options...)
}

func langchainCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []langchain.Option

	return langchain.NewCompleter(cfg.URL, options...)
}

func llamaCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []llama.Option

	if model.ID != "" {
		options = append(options, llama.WithModel(model.ID))
	}

	return llama.NewCompleter(cfg.URL, options...)
}

func mistralCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []mistral.Option

	if cfg.Token != "" {
		options = append(options, mistral.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, mistral.WithModel(model.ID))
	}

	return mistral.NewCompleter(options...)
}

func ollamaCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []ollama.Option

	if model.ID != "" {
		options = append(options, ollama.WithModel(model.ID))
	}

	return ollama.NewCompleter(cfg.URL, options...)
}

func openaiCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []openai.Option

	if cfg.URL != "" {
		options = append(options, openai.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, openai.WithToken(cfg.Token))
	}

	if model.ID != "" {
		options = append(options, openai.WithModel(model.ID))
	}

	return openai.NewCompleter(options...)
}

func customCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []custom.Option

	return custom.NewCompleter(cfg.URL, options...)
}
