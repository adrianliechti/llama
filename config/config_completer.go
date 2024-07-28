package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/adapter"
	"github.com/adrianliechti/llama/pkg/adapter/hermesfn"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/anthropic"
	"github.com/adrianliechti/llama/pkg/provider/custom"
	"github.com/adrianliechti/llama/pkg/provider/groq"
	"github.com/adrianliechti/llama/pkg/provider/huggingface"
	"github.com/adrianliechti/llama/pkg/provider/langchain"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/mistral"
	"github.com/adrianliechti/llama/pkg/provider/ollama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

func (cfg *Config) RegisterCompleter(model string, c provider.Completer) {
	cfg.RegisterModel(model)

	if cfg.completer == nil {
		cfg.completer = make(map[string]provider.Completer)
	}

	cfg.completer[model] = c
}

func createCompleter(cfg providerConfig, model string) (provider.Completer, error) {
	switch strings.ToLower(cfg.Type) {

	case "anthropic":
		return anthropicCompleter(cfg, model)

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

func createCompleterAdapter(name string, completer provider.Completer) (adapter.Provider, error) {
	switch strings.ToLower(name) {

	case "hermesfn", "hermes-function-calling":
		return hermesfn.New(completer)

	default:
		return nil, errors.New("invalid adapter type: " + name)
	}
}

func anthropicCompleter(cfg providerConfig, model string) (provider.Completer, error) {
	var options []anthropic.Option

	if cfg.URL != "" {
		options = append(options, anthropic.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, anthropic.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, anthropic.WithModel(model))
	}

	return anthropic.NewCompleter(options...)
}

func groqCompleter(cfg providerConfig, model string) (provider.Completer, error) {
	var options []groq.Option

	if cfg.Token != "" {
		options = append(options, groq.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, groq.WithModel(model))
	}

	return groq.NewCompleter(options...)
}

func huggingfaceCompleter(cfg providerConfig, model string) (provider.Completer, error) {
	var options []huggingface.Option

	if cfg.Token != "" {
		options = append(options, huggingface.WithToken(cfg.Token))
	}

	return huggingface.NewCompleter(cfg.URL, options...)
}

func langchainCompleter(cfg providerConfig, model string) (provider.Completer, error) {
	var options []langchain.Option

	return langchain.NewCompleter(cfg.URL, options...)
}

func llamaCompleter(cfg providerConfig, model string) (provider.Completer, error) {
	var options []llama.Option

	if model != "" {
		options = append(options, llama.WithModel(model))
	}

	return llama.NewCompleter(cfg.URL, options...)
}

func mistralCompleter(cfg providerConfig, model string) (provider.Completer, error) {
	var options []mistral.Option

	if cfg.Token != "" {
		options = append(options, mistral.WithToken(cfg.Token))
	}

	if model != "" {
		options = append(options, mistral.WithModel(model))
	}

	return mistral.NewCompleter(options...)
}

func ollamaCompleter(cfg providerConfig, model string) (provider.Completer, error) {
	var options []ollama.Option

	if model != "" {
		options = append(options, ollama.WithModel(model))
	}

	return ollama.NewCompleter(cfg.URL, options...)
}

func openaiCompleter(cfg providerConfig, model string) (provider.Completer, error) {
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

	return openai.NewCompleter(options...)
}

func customCompleter(cfg providerConfig, model string) (provider.Completer, error) {
	var options []custom.Option

	return custom.NewCompleter(cfg.URL, options...)
}
