package config

import (
	"errors"
	"sort"

	"github.com/adrianliechti/llama/pkg/authorizer"
	"github.com/adrianliechti/llama/pkg/chain"
	"github.com/adrianliechti/llama/pkg/classifier"
	"github.com/adrianliechti/llama/pkg/extractor"
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
	synthesizer map[string]provider.Synthesizer
	translator  map[string]provider.Translator
	transcriber map[string]provider.Transcriber
	renderer    map[string]provider.Renderer

	indexes     map[string]index.Provider
	extractors  map[string]extractor.Provider
	classifiers map[string]classifier.Provider

	tools  map[string]tool.Tool
	chains map[string]chain.Provider
}

func (cfg *Config) Models() []provider.Model {
	var result []provider.Model

	for _, m := range cfg.models {
		result = append(result, m)
	}

	sort.SliceStable(result, func(i, j int) bool { return result[i].ID < result[j].ID })

	return result
}

func (cfg *Config) Model(id string) (*provider.Model, error) {
	if cfg.models != nil {
		if m, ok := cfg.models[id]; ok {
			return &m, nil
		}
	}

	return nil, errors.New("model not found: " + id)
}

func (cfg *Config) Embedder(model string) (provider.Embedder, error) {
	if cfg.embedder != nil {
		if e, ok := cfg.embedder[model]; ok {
			return e, nil
		}
	}

	return nil, errors.New("embedder not found: " + model)
}

func (cfg *Config) Completer(model string) (provider.Completer, error) {
	if cfg.completer != nil {
		if c, ok := cfg.completer[model]; ok {
			return c, nil
		}
	}

	if cfg.chains != nil {
		if c, ok := cfg.chains[model]; ok {
			return c, nil
		}
	}

	return nil, errors.New("completer not found: " + model)
}

func (cfg *Config) Synthesizer(model string) (provider.Synthesizer, error) {
	if cfg.synthesizer != nil {
		if s, ok := cfg.synthesizer[model]; ok {
			return s, nil
		}
	}

	return nil, errors.New("synthesizer not found: " + model)
}

func (cfg *Config) Translator(model string) (provider.Translator, error) {
	if cfg.translator != nil {
		if t, ok := cfg.translator[model]; ok {
			return t, nil
		}
	}

	return nil, errors.New("translator not found: " + model)
}

func (cfg *Config) Transcriber(model string) (provider.Transcriber, error) {
	if cfg.transcriber != nil {
		if t, ok := cfg.transcriber[model]; ok {
			return t, nil
		}
	}

	return nil, errors.New("transcriber not found: " + model)
}

func (cfg *Config) Renderer(model string) (provider.Renderer, error) {
	if cfg.renderer != nil {
		if t, ok := cfg.renderer[model]; ok {
			return t, nil
		}
	}

	return nil, errors.New("renderer not found: " + model)
}

func (cfg *Config) Index(id string) (index.Provider, error) {
	if cfg.indexes != nil {
		if i, ok := cfg.indexes[id]; ok {
			return i, nil
		}
	}

	return nil, errors.New("index not found: " + id)
}

func (cfg *Config) Extractor(id string) (extractor.Provider, error) {
	if cfg.extractors != nil {
		if e, ok := cfg.extractors[id]; ok {
			return e, nil
		}
	}

	return nil, errors.New("extractor not found: " + id)
}

func (cfg *Config) Tool(id string) (tool.Tool, error) {
	if cfg.tools != nil {
		if t, ok := cfg.tools[id]; ok {
			return t, nil
		}
	}

	return nil, errors.New("tool not found: " + id)
}

func (cfg *Config) Classifier(id string) (classifier.Provider, error) {
	if cfg.classifiers != nil {
		if c, ok := cfg.classifiers[id]; ok {
			return c, nil
		}
	}

	return nil, errors.New("classifier not found: " + id)
}

func Parse(path string) (*Config, error) {
	file, err := parseFile(path)

	if err != nil {
		return nil, err
	}

	c := &Config{
		Address: ":8080",
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

	if err := c.registerExtractors(file); err != nil {
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
