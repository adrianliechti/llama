package config

import (
	"errors"
	"os"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
)

var (
	ErrModelNotFound     = errors.New("model not found")
	ErrEmbedderNotFound  = errors.New("embedder not found")
	ErrCompleterNotFound = errors.New("completer not found")
)

type Config struct {
	Address string

	Authorizers []authorizer.Provider

	models map[string]Model

	providers map[string]provider.Provider
	indexes   map[string]index.Provider
	chains    map[string]chain.Provider
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

func (c *Config) Model(id string) (*Model, error) {
	m, ok := c.models[id]

	if !ok {
		return nil, ErrModelNotFound
	}

	return &m, nil
}

func (c *Config) Embedder(model string) (provider.Embedder, error) {
	m, err := c.Model(model)

	if err != nil {
		return nil, err
	}

	if p, ok := c.providers[model]; ok {
		return provider.ToEmbbedder(p, m.model), nil
	}

	return nil, ErrEmbedderNotFound
}

func (c *Config) Completer(model string) (provider.Completer, error) {
	m, err := c.Model(model)

	if err != nil {
		return nil, err
	}

	if p, ok := c.providers[model]; ok {
		return provider.ToCompleter(p, m.model), nil
	}

	if c, ok := c.chains[model]; ok {
		return c, nil
	}

	return nil, ErrCompleterNotFound
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

		providers: make(map[string]provider.Provider),
		indexes:   make(map[string]index.Provider),
		chains:    make(map[string]chain.Provider),
	}

	if err := c.registerAuthorizer(file); err != nil {
		return nil, err
	}

	if err := c.registerProviders(file); err != nil {
		return nil, err
	}

	if err := c.registerChains(file); err != nil {
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
