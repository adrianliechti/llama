package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
)

func createProvider(p providerConfig) (provider.Provider, error) {
	switch strings.ToLower(p.Type) {
	case "openai":
		return openaiProvider(p)

	case "llama":
		return llamaProvider(p)

	default:
		return nil, errors.New("invalid provider type: " + p.Type)
	}
}

func openaiProvider(p providerConfig) (provider.Provider, error) {
	var options []openai.Option

	if p.URL != "" {
		options = append(options, openai.WithURL(p.URL))
	}

	if p.Token != "" {
		options = append(options, openai.WithToken(p.Token))
	}

	models := p.Models

	if len(models) > 0 {
		var mapper modelMapper = models

		options = append(options, openai.WithModelMapper(mapper))
	}

	return openai.New(options...), nil
}

func llamaProvider(p providerConfig) (provider.Provider, error) {
	var options []llama.Option

	if p.URL != "" {
		options = append(options, llama.WithURL(p.URL))
	}

	if len(p.Models) > 1 {
		return nil, errors.New("multiple models not supported for llama provider")
	}

	var model string
	var prompt string
	var template string

	for k, v := range p.Models {
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
		options = append(options, llama.WithPromptTemplate(&llama.PromptTemplateChatML{}))

	case "llama":
		options = append(options, llama.WithPromptTemplate(&llama.PromptTemplateLLAMA{}))

	case "mistral":
		options = append(options, llama.WithPromptTemplate(&llama.PromptTemplateMistral{}))

	default:
		return nil, errors.New("invalid prompt template: " + template)
	}

	return llama.New(options...), nil
}
