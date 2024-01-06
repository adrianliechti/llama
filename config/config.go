package config

import (
	"errors"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/provider"
)

var (
	ErrModelNotFound = errors.New("model not found")

	ErrEmbedderNotFound    = errors.New("embedder not found")
	ErrCompleterNotFound   = errors.New("completer not found")
	ErrTranscriberNotFound = errors.New("transcriber not found")

	ErrIndexNotFound      = errors.New("index not found")
	ErrClassifierNotFound = errors.New("classifier not found")
)

type Config struct {
	Address string

	Authorizers []authorizer.Provider

	models map[string]provider.Model

	embedder    map[string]provider.Embedder
	completer   map[string]provider.Completer
	transcriber map[string]provider.Transcriber

	indexes     map[string]index.Provider
	classifiers map[string]classifier.Provider

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
		return nil, ErrModelNotFound
	}

	return &m, nil
}

func (cfg *Config) Embedder(model string) (provider.Embedder, error) {
	if e, ok := cfg.embedder[model]; ok {
		return e, nil
	}

	return nil, ErrEmbedderNotFound
}

func (cfg *Config) Completer(model string) (provider.Completer, error) {
	if c, ok := cfg.completer[model]; ok {
		return c, nil
	}

	if c, ok := cfg.chains[model]; ok {
		return c, nil
	}

	return nil, ErrCompleterNotFound
}

func (cfg *Config) Transcriber(model string) (provider.Transcriber, error) {
	if c, ok := cfg.transcriber[model]; ok {
		return c, nil
	}

	return nil, ErrTranscriberNotFound
}

func (cfg *Config) Index(id string) (index.Provider, error) {
	i, ok := cfg.indexes[id]

	if !ok {
		return nil, ErrIndexNotFound
	}

	return i, nil
}

func (cfg *Config) Classifier(id string) (classifier.Provider, error) {
	c, ok := cfg.classifiers[id]

	if !ok {
		return nil, ErrClassifierNotFound
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
		classifiers: make(map[string]classifier.Provider),

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

	if err := c.registerClassifiers(file); err != nil {
		return nil, err
	}

	if err := c.registerChains(file); err != nil {
		return nil, err
	}

	return c, nil
}
