package config

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/authorizer/oidc"
	"github.com/adrianliechti/llama/pkg/authorizer/static"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/llama"
	"github.com/adrianliechti/llama/pkg/provider/openai"
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

	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var config configFile

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	addr := addrFromEnvironment()

	providers, err := providersFromConfig(config.Providers)

	if err != nil {
		return nil, err
	}

	authorizer, err := authorizerFromConfig(config.Auth)

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
		case "", "openai":
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

func authorizerFromConfig(auth authConfig) (authorizer.Provider, error) {
	if auth.Issuer != "" {
		return oidc.New(auth.Issuer, auth.Audience)
	}

	if auth.Token != "" {
		return static.New(auth.Token)
	}

	return nil, nil
}

type configFile struct {
	Auth authConfig `yaml:"auth"`

	Providers []providerConfig `yaml:"providers"`
}

type authConfig struct {
	Token string `yaml:"token"`

	Issuer   string `yaml:"issuer"`
	Audience string `yaml:"audience"`
}

type providerConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	Models map[string]modelConfig `yaml:"models"`
}

type modelConfig struct {
	ID string `yaml:"id"`

	Prompt   string `yaml:"prompt"`
	Template string `yaml:"template"`

	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type modelMapper map[string]modelConfig

func (m modelMapper) From(val string) string {
	for k, v := range m {
		if v.ID != "" && strings.EqualFold(v.ID, val) {
			return k
		}
	}

	for k := range m {
		if strings.EqualFold(k, val) {
			return k
		}
	}

	return ""
}

func (m modelMapper) To(val string) string {
	for k, v := range m {
		if strings.EqualFold(k, val) {
			if v.ID != "" {
				return v.ID
			}

			return k
		}
	}

	return ""
}
