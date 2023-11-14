package config

import (
	"errors"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/adrianliechti/llama/pkg/auth"
	"github.com/adrianliechti/llama/pkg/llm/dispatcher"
	"github.com/adrianliechti/llama/pkg/llm/llama"
	"github.com/adrianliechti/llama/pkg/llm/openai"

	"github.com/adrianliechti/llama/pkg/llm"
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

	llm, err := llmFromConfig(config.Providers)

	if err != nil {
		return nil, err
	}

	return &Config{
		Addr: addr,

		LLM: llm,
	}, nil
}

func addrFromEnvironment() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return ":" + port
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

			llm := llama.New(options...)
			llms = append(llms, llm)

		default:
			return nil, errors.New("invalid provider type: " + p.Type)
		}
	}

	return dispatcher.New(llms...)
}

type configFile struct {
	Providers []providerConfig `yaml:"providers"`
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
