package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

func createProvider(c providerConfig) (provider.Provider, error) {
	switch strings.ToLower(c.Type) {
	case "openai":
		return openaiProvider(c)

	case "llama":
		return llamaProvider(c)

	default:
		return nil, errors.New("invalid provider type: " + c.Type)
	}
}

func openaiProvider(c providerConfig) (provider.Provider, error) {
	var options []openai.Option

	if c.URL != "" {
		options = append(options, openai.WithURL(c.URL))
	}

	if c.Token != "" {
		options = append(options, openai.WithToken(c.Token))
	}

	models := c.Models

	if len(models) > 0 {
		var mapper modelMapper = models

		options = append(options, openai.WithModelMapper(mapper))
	}

	return openai.New(options...), nil
}

func llamaProvider(c providerConfig) (provider.Provider, error) {
	var options []llama.Option

	if c.URL != "" {
		options = append(options, llama.WithURL(c.URL))
	}

	if len(c.Models) > 1 {
		return nil, errors.New("multiple models not supported for llama provider")
	}

	var model string
	var prompt string
	var template string

	for k, v := range c.Models {
		model = k
		prompt = v.Prompt
		template = v.Template

		break
	}

	if model != "" {
		options = append(options, llama.WithModel(model))
	}

	if prompt != "" {
		options = append(options, llama.WithSystem(prompt))
	}

	switch strings.ToLower(template) {
	case "chatml":
		options = append(options, llama.WithPromptTemplate(&llama.PromptChatML{}))

	case "llama":
		options = append(options, llama.WithPromptTemplate(&llama.PromptLLAMA{}))

	case "mistral":
		options = append(options, llama.WithPromptTemplate(&llama.PromptMistral{}))

	default:
		return nil, errors.New("invalid prompt template: " + template)
	}

	return llama.New(options...), nil
}
