package config

import (
	"errors"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/extracter"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/tool"
)

type Config struct {
	Address string

	Authorizers []authorizer.Provider

	models map[string]provider.Model

	embedder    map[string]provider.Embedder
	completer   map[string]provider.Completer
	transcriber map[string]provider.Transcriber

	indexes     map[string]index.Provider
	extracters  map[string]extracter.Provider
	classifiers map[string]classifier.Provider

	tools  map[string]tool.Tool
	chains map[string]chain.Provider
}

func (cfg *Config) Models() []provider.Model {
	var result []provider.Model

	for _, m := range cfg.models {
		result = append(result, m)
	}

	return result
}

func (cfg *Config) Model(id string) (*provider.Model, error) {
	m, ok := cfg.models[id]

	if !ok {
		return nil, errors.New("model not found: " + id)
	}

	return &m, nil
}

func (cfg *Config) Embedder(model string) (provider.Embedder, error) {
	if e, ok := cfg.embedder[model]; ok {
		return e, nil
	}

	return nil, errors.New("embedder not found: " + model)
}

func (cfg *Config) Completer(model string) (provider.Completer, error) {
	if c, ok := cfg.completer[model]; ok {
		return c, nil
	}

	if c, ok := cfg.chains[model]; ok {
		return c, nil
	}

	return nil, errors.New("completer not found: " + model)
}

func (cfg *Config) Transcriber(model string) (provider.Transcriber, error) {
	if c, ok := cfg.transcriber[model]; ok {
		return c, nil
	}

	return nil, errors.New("transcriber not found: " + model)
}

func (cfg *Config) Index(id string) (index.Provider, error) {
	i, ok := cfg.indexes[id]

	if !ok {
		return nil, errors.New("index not found: " + id)
	}

	return i, nil
}

func (cfg *Config) Extracter(id string) (extracter.Provider, error) {
	e, ok := cfg.extracters[id]

	if !ok {
		return nil, errors.New("extracter not found: " + id)
	}

	return e, nil
}

func (cfg *Config) Tool(id string) (tool.Tool, error) {
	t, ok := cfg.tools[id]

	if !ok {
		return nil, errors.New("tool not found: " + id)
	}

	return t, nil
}

func (cfg *Config) Classifier(id string) (classifier.Provider, error) {
	c, ok := cfg.classifiers[id]

	if !ok {
		return nil, errors.New("classifier not found: " + id)
	}

	return c, nil
}

func Parse(path string) (*Config, error) {
	file, err := parseFile(path)

	if err != nil {
		return nil, err
	}

	c := &Config{
		Address: ":8080",

		models: make(map[string]provider.Model),

		embedder:    make(map[string]provider.Embedder),
		completer:   make(map[string]provider.Completer),
		transcriber: make(map[string]provider.Transcriber),

		indexes:     make(map[string]index.Provider),
		extracters:  make(map[string]extracter.Provider),
		classifiers: make(map[string]classifier.Provider),

		tools:  make(map[string]tool.Tool),
		chains: make(map[string]chain.Provider),
	}

	if err := c.registerAuthorizer(file); err != nil {
		return nil, err
	}

	if err := c.registerProviders(file); err != nil {
		return nil, err
	}

	if err := c.registerIndexes(file); err != nil {
		return nil, err
	}

	if err := c.registerExtracters(file); err != nil {
		return nil, err
	}

	if err := c.registerClassifiers(file); err != nil {
		return nil, err
	}

	if err := c.registerTools(file); err != nil {
		return nil, err
	}

	if err := c.registerChains(file); err != nil {
		return nil, err
	}

	return c, nil
}
