package config

import (
	"errors"
	"os"
	"strings"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/authorizer/oidc"
	"github.com/adrianliechti/llama/pkg/authorizer/static"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/openai"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/index/chroma"

	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/chain/rag"
)

type Config struct {
	Addr string

	Providers  []provider.Provider
	Authorizer authorizer.Provider
}

func Parse(path string) (*Config, error) {
	if path == "" {
		path = "config.yaml"
	}

	file, err := parseFile(path)

	if err != nil {
		return nil, err
	}

	addr := addrFromEnvironment()

	providers, err := providersFromConfig(file.Providers)

	if err != nil {
		return nil, err
	}

	chains, err := RAGsFromConfig(file.Chains)

	if err != nil {
		return nil, err
	}

	_ = chains

	authorizer, err := authorizerFromConfig(file.Auth)

	if err != nil {
		return nil, err
	}

	return &Config{
		Addr: addr,

		Providers:  providers,
		Authorizer: authorizer,
	}, nil
}

func addrFromEnvironment() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return ":" + port
}

func providersFromConfig(providers []providerConfig) ([]provider.Provider, error) {
	var result []provider.Provider

	for _, p := range providers {
		switch strings.ToLower(p.Type) {
		case "openai":
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

			p := openai.New(options...)
			result = append(result, p)

		case "llama":
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

			p := llama.New(options...)
			result = append(result, p)

		default:
			return nil, errors.New("invalid provider type: " + p.Type)
		}
	}

	return result, nil
}

func RAGsFromConfig(chains map[string]chainConfig) ([]chain.Provider, error) {
	var result []chain.Provider

	for id, c := range chains {
		if !strings.EqualFold(c.Type, "rag") {
			continue
		}

		_ = id

		var index index.Index

		if c.Index != nil {
			switch strings.ToLower(c.Index.Type) {
			case "chroma":
				i, err := chroma.New(c.Index.URL, c.Index.Name)

				if err != nil {
					return nil, err
				}

				index = i

			default:
				return nil, errors.New("invalid index type: " + c.Index.Type)
			}
		}

		switch strings.ToLower(c.Type) {
		case "rag":
			rag := rag.New(index, nil, nil)

			result = append(result, rag)
		}
	}

	return result, nil
}

func authorizerFromConfig(auth authConfig) (authorizer.Provider, error) {
	if auth.Issuer != "" {
		return oidc.New(auth.Issuer, auth.Audience)
	}

	if auth.Token != "" {
		return static.New(auth.Token)
	}

	return nil, nil
}
