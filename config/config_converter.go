package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/converter"
	"github.com/adrianliechti/llama/pkg/converter/azure"
	"github.com/adrianliechti/llama/pkg/converter/multi"
	"github.com/adrianliechti/llama/pkg/converter/tika"
	"github.com/adrianliechti/llama/pkg/converter/unstructured"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (cfg *Config) RegisterConverter(name, alias string, p converter.Provider) {
	if cfg.converters == nil {
		cfg.converters = make(map[string]converter.Provider)
	}

	converter, ok := p.(otel.ObservableConverter)

	if !ok {
		converter = otel.NewConverter(name, p)
	}

	cfg.converters[alias] = converter
}

func (cfg *Config) Converter(id string) (converter.Provider, error) {
	if cfg.converters != nil {
		if c, ok := cfg.converters[id]; ok {
			return c, nil
		}
	}

	return nil, errors.New("converter not found: " + id)
}

type converterConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

func (cfg *Config) RegisterConverters(f *configFile) error {
	var converters []converter.Provider

	for id, c := range f.Converters {
		converter, err := createConverter(c)

		if err != nil {
			return err
		}

		converters = append(converters, converter)

		cfg.RegisterConverter(c.Type, id, converter)
	}

	cfg.RegisterConverter("default", "default", multi.New(converters...))

	return nil
}

func createConverter(cfg converterConfig) (converter.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "azure":
		return azureConverter(cfg)

	case "tika":
		return tikaConverter(cfg)

	case "unstructured":
		return unstructuredConverter(cfg)

	default:
		return nil, errors.New("invalid converter type: " + cfg.Type)
	}
}

func azureConverter(cfg converterConfig) (converter.Provider, error) {
	var options []azure.Option

	if cfg.Token != "" {
		options = append(options, azure.WithToken(cfg.Token))
	}

	return azure.New(cfg.URL, options...)
}

func tikaConverter(cfg converterConfig) (converter.Provider, error) {
	var options []tika.Option

	return tika.New(cfg.URL, options...)
}

func unstructuredConverter(cfg converterConfig) (converter.Provider, error) {
	var options []unstructured.Option

	if cfg.Token != "" {
		options = append(options, unstructured.WithToken(cfg.Token))
	}

	return unstructured.New(cfg.URL, options...)
}
