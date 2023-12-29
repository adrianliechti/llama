package config

import (
	"os"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
)

type Config struct {
	Address string

	Authorizers []authorizer.Provider

	models map[string]Model

	indexes   map[string]index.Provider
	providers map[string]provider.Provider
}

type Model struct {
	ID string

	model string
}

func (c *Config) Models() []Model {
	var result []Model

	for _, m := range c.models {
		result = append(result, m)
	}

	return result
}

func (c *Config) Model(id string) (Model, bool) {
	m, ok := c.models[id]
	return m, ok
}

func (c *Config) Embedder(model string) (provider.Embedder, bool) {
	m, found := c.Model(model)

	if !found {
		return nil, false
	}

	p, found := c.providers[model]

	if !found {
		return nil, false
	}

	return provider.ToEmbbedder(p, m.model), true
}

func (c *Config) Completer(model string) (provider.Completer, bool) {
	m, found := c.Model(model)

	if !found {
		return nil, false
	}

	p, ok := c.providers[model]

	if !ok {
		return nil, false
	}

	return provider.ToCompleter(p, m.model), true
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

		models: make(map[string]Model),

		indexes:   make(map[string]index.Provider),
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

		c.Authorizers = append(c.Authorizers, authorizer)
	}

	return nil
}

func (c *Config) registerChains(f *configFile) error {
	for id, cfg := range f.Chains {

		if cfg.Model != "" {
		}

		if cfg.Embedding != "" {
		}

		if cfg.Index != nil {
			i, err := createIndex(*cfg.Index)

			if err != nil {
				return err
			}

			c.indexes[id] = i
		}
	}

	return nil
}

func (c *Config) registerProviders(f *configFile) error {
	for _, cfg := range f.Providers {
		p, err := createProvider(cfg)

		if err != nil {
			return err
		}

		for id, cfg := range cfg.Models {
			c.models[id] = Model{
				ID: id,

				model: cfg.ID,
			}

			c.providers[id] = p
		}
	}

	return nil
}
