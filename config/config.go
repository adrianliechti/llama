package config

import (
	"os"

	"github.com/adrianliechti/llama/pkg/authorizer"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Config struct {
	Address string

	Authorizer []authorizer.Provider
	Providers  []provider.Provider

	models    map[string]provider.Model
	providers map[string]provider.Provider
}

func (c *Config) Models() []provider.Model {
	var result []provider.Model

	for _, m := range c.models {
		result = append(result, m)
	}

	return result
}

func (c *Config) Model(id string) (provider.Model, bool) {
	m, ok := c.models[id]
	return m, ok
}

func (c *Config) Provider(model string) (provider.Provider, bool) {
	p, ok := c.providers[model]
	return p, ok
}

func Parse(path string) (*Config, error) {
	if path == "" {
		path = "config.yaml"
	}

	file, err := parseFile(path)

	if err != nil {
		return nil, err
	}

	c := &Config{
		Address: addrFromEnvironment(),

		Authorizer: make([]authorizer.Provider, 0),
		Providers:  make([]provider.Provider, 0),

		models:    make(map[string]provider.Model),
		providers: make(map[string]provider.Provider),
	}

	if err := c.registerAuthorizer(file); err != nil {
		return nil, err
	}

	if err := c.registerProviders(file); err != nil {
		return nil, err
	}

	return c, nil
}

func addrFromEnvironment() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return ":" + port
}

func (c *Config) registerAuthorizer(f *configFile) error {
	for _, a := range f.Authorizers {
		authorizer, err := createAuthorizer(a)

		if err != nil {
			return err
		}

		c.Authorizer = append(c.Authorizer, authorizer)
	}

	return nil
}

func (c *Config) registerProviders(f *configFile) error {
	for _, cfg := range f.Providers {
		p, err := createProvider(cfg)

		if err != nil {
			return err
		}

		for model := range cfg.Models {
			c.models[model] = provider.Model{
				ID: model,
			}

			c.providers[model] = p
		}

		c.Providers = append(c.Providers, p)
	}

	return nil
}

// func RAGsFromConfig(chains map[string]chainConfig) ([]chain.Provider, error) {
// 	var result []chain.Provider

// 	for id, c := range chains {
// 		if !strings.EqualFold(c.Type, "rag") {
// 			continue
// 		}

// 		_ = id

// 		var index index.Index

// 		if c.Index != nil {
// 			switch strings.ToLower(c.Index.Type) {
// 			case "chroma":
// 				i, err := chroma.New(c.Index.URL, c.Index.Name)

// 				if err != nil {
// 					return nil, err
// 				}

// 				index = i

// 			default:
// 				return nil, errors.New("invalid index type: " + c.Index.Type)
// 			}
// 		}

// 		switch strings.ToLower(c.Type) {
// 		case "rag":
// 			rag := rag.New(index, nil, nil)

// 			result = append(result, rag)
// 		}
// 	}

// 	return result, nil
// }
