package config

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/adrianliechti/llama/pkg/auth"
	"github.com/adrianliechti/llama/pkg/auth/oidc"
	"github.com/adrianliechti/llama/pkg/auth/static"

	"github.com/adrianliechti/llama/pkg/llm"
	"github.com/adrianliechti/llama/pkg/llm/dispatcher"
	"github.com/adrianliechti/llama/pkg/llm/llama"
	"github.com/adrianliechti/llama/pkg/llm/openai"
)

type Config struct {
	Addr string

	Auth auth.Provider
	LLM  llm.Provider
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

	auth, err := authFromConfig(config.Auth)

	if err != nil {
		return nil, err
	}

	llm, err := llmFromConfig(config.Providers)

	if err != nil {
		return nil, err
	}

	return &Config{
		Addr: addr,

		Auth: auth,
		LLM:  llm,
	}, nil
}

func addrFromEnvironment() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return ":" + port
}

func authFromConfig(auth authConfig) (auth.Provider, error) {
	if auth.Issuer != "" {
		return oidc.New(auth.Issuer, auth.Audience)
	}

	if auth.Token != "" {
		return static.New(auth.Token)
	}

	return nil, nil
}

func llmFromConfig(providers []providerConfig) (llm.Provider, error) {
	var llms []llm.Provider

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
				var mapper openai.ModelMapper = func(model string) string {
					for _, m := range models {
						if strings.EqualFold(m.ID, model) {
							if m.Alias != "" {
								return m.Alias
							}

							return m.ID
						}
					}

					return ""
				}

				options = append(options, openai.WithModelMapper(mapper))
			}

			llm := openai.New(options...)
			llms = append(llms, llm)

		case "llama":
			var options []llama.Option

			if p.URL != "" {
				options = append(options, llama.WithURL(p.URL))
			}

			if len(p.Models) > 0 {
				options = append(options, llama.WithModel(p.Models[0].ID))
			}

			options = append(options, llama.WithPromptTemplate(&llama.PromptTemplateLLAMA{}))
			options = append(options, llama.WithSystem("You are a helpful, respectful and honest assistant. Always answer as helpfully as possible, while being safe. Your answers should not include any harmful, unethical, racist, sexist, toxic, dangerous, or illegal content. Please ensure that your responses are socially unbiased and positive in nature.\n\nIf a question does not make any sense, or is not factually coherent, explain why instead of answering something not correct. If you don't know the answer to a question, please don't share false information."))

			llm := llama.New(options...)
			llms = append(llms, llm)

		case "orca":
			var options []llama.Option

			if p.URL != "" {
				options = append(options, llama.WithURL(p.URL))
			}

			if len(p.Models) > 0 {
				options = append(options, llama.WithModel(p.Models[0].ID))
			}

			options = append(options, llama.WithPromptTemplate(&llama.PromptTemplateChatML{}))
			options = append(options, llama.WithSystem("You are Orca, an AI language model created by Microsoft. You are a cautious assistant. You carefully follow instructions. You are helpful and harmless and you follow ethical guidelines and promote positive behavior."))

			llm := llama.New(options...)
			llms = append(llms, llm)

		default:
			return nil, errors.New("invalid provider type: " + p.Type)
		}
	}

	return dispatcher.New(llms...)
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

	Models []modelConfig `yaml:"models"`
}

type modelConfig struct {
	ID string `yaml:"id"`

	Name  string `yaml:"name"`
	Alias string `yaml:"alias"`
}
