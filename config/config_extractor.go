package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/extractor/azure"
	"github.com/adrianliechti/llama/pkg/extractor/jina"
	"github.com/adrianliechti/llama/pkg/extractor/multi"
	"github.com/adrianliechti/llama/pkg/extractor/text"
	"github.com/adrianliechti/llama/pkg/extractor/tika"
	"github.com/adrianliechti/llama/pkg/extractor/unstructured"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (cfg *Config) RegisterExtractor(name, alias string, p extractor.Provider) {
	if cfg.extractors == nil {
		cfg.extractors = make(map[string]extractor.Provider)
	}

	extractor, ok := p.(otel.ObservableExtractor)

	if !ok {
		extractor = otel.NewExtractor(name, p)
	}

	cfg.extractors[alias] = extractor
}

func (cfg *Config) Extractor(id string) (extractor.Provider, error) {
	if cfg.extractors != nil {
		if c, ok := cfg.extractors[id]; ok {
			return c, nil
		}
	}

	return nil, errors.New("extractor not found: " + id)
}

type extractorConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

func (cfg *Config) RegisterExtractors(f *configFile) error {
	var extractors []extractor.Provider

	for id, c := range f.Extractors {
		extractor, err := createExtractor(c)

		if err != nil {
			return err
		}

		extractors = append(extractors, extractor)

		cfg.RegisterExtractor(c.Type, id, extractor)
	}

	cfg.RegisterExtractor("default", "default", multi.New(extractors...))

	return nil
}

func createExtractor(cfg extractorConfig) (extractor.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "azure":
		return azureExtractor(cfg)

	case "jina":
		return jinaExtractor(cfg)

	case "text":
		return textExtractor(cfg)

	case "tika":
		return tikaExtractor(cfg)

	case "unstructured":
		return unstructuredExtractor(cfg)

	default:
		return nil, errors.New("invalid extractor type: " + cfg.Type)
	}
}

func azureExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []azure.Option

	if cfg.Token != "" {
		options = append(options, azure.WithToken(cfg.Token))
	}

	return azure.New(cfg.URL, options...)
}

func jinaExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []jina.Option

	if cfg.Token != "" {
		options = append(options, jina.WithToken(cfg.Token))
	}

	return jina.New(cfg.URL, options...)
}

func textExtractor(cfg extractorConfig) (extractor.Provider, error) {
	return text.New()
}

func tikaExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []tika.Option

	return tika.New(cfg.URL, options...)
}

func unstructuredExtractor(cfg extractorConfig) (extractor.Provider, error) {
	var options []unstructured.Option

	if cfg.Token != "" {
		options = append(options, unstructured.WithToken(cfg.Token))
	}

	return unstructured.New(cfg.URL, options...)
}
