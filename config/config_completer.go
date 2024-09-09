package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/otel"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/anthropic"
	"github.com/adrianliechti/llama/pkg/provider/azure"
	"github.com/adrianliechti/llama/pkg/provider/cohere"
	"github.com/adrianliechti/llama/pkg/provider/custom"
	"github.com/adrianliechti/llama/pkg/provider/groq"
	"github.com/adrianliechti/llama/pkg/provider/huggingface"
	"github.com/adrianliechti/llama/pkg/provider/langchain"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/mistral"
	"github.com/adrianliechti/llama/pkg/provider/mistralrs"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
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

	case "azure":
		return azureCompleter(cfg, model)

	case "cohere":
		return cohereCompleter(cfg, model)

	case "github":
		return azureCompleter(cfg, model)

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

	case "mistralrs":
		return mistralrsCompleter(cfg, model)

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

	return anthropic.NewCompleter(model.ID, options...)
}

func azureCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []azure.Option

	if cfg.URL != "" {
		options = append(options, azure.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, azure.WithToken(cfg.Token))
	}

	return azure.NewCompleter(model.ID, options...)
}

func cohereCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []cohere.Option

	if cfg.Token != "" {
		options = append(options, cohere.WithToken(cfg.Token))
	}

	return cohere.NewCompleter(model.ID, options...)
}

func groqCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []groq.Option

	if cfg.Token != "" {
		options = append(options, groq.WithToken(cfg.Token))
	}

	return groq.NewCompleter(model.ID, options...)
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

	return llama.NewCompleter(model.ID, cfg.URL, options...)
}

func mistralCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []mistral.Option

	if cfg.Token != "" {
		options = append(options, mistral.WithToken(cfg.Token))
	}

	return mistral.NewCompleter(model.ID, options...)
}

func mistralrsCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []mistralrs.Option

	return mistralrs.NewCompleter(cfg.URL, model.ID, options...)
}

func ollamaCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []ollama.Option

	return ollama.NewCompleter(cfg.URL, model.ID, options...)
}

func openaiCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []openai.Option

	if cfg.URL != "" {
		options = append(options, openai.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, openai.WithToken(cfg.Token))
	}

	if model.Limiter != nil {
		options = append(options, openai.WithLimiter(model.Limiter))
	}

	return openai.NewCompleter(model.ID, options...)
}

func customCompleter(cfg providerConfig, model modelContext) (provider.Completer, error) {
	var options []custom.Option

	return custom.NewCompleter(cfg.URL, options...)
}
