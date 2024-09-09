package config

import (
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/partitioner"
	"github.com/adrianliechti/llama/pkg/partitioner/azure"
	"github.com/adrianliechti/llama/pkg/partitioner/code"
	"github.com/adrianliechti/llama/pkg/partitioner/multi"
	"github.com/adrianliechti/llama/pkg/partitioner/text"
	"github.com/adrianliechti/llama/pkg/partitioner/tika"
	"github.com/adrianliechti/llama/pkg/partitioner/unstructured"

	"github.com/adrianliechti/llama/pkg/otel"
)

func (cfg *Config) RegisterPartitioner(name, alias string, p partitioner.Provider) {
	if cfg.partitioners == nil {
		cfg.partitioners = make(map[string]partitioner.Provider)
	}

	partitioner, ok := p.(otel.ObservablePartitioner)

	if !ok {
		partitioner = otel.NewPartitioner(name, p)
	}

	cfg.partitioners[alias] = partitioner
}

func (cfg *Config) Partitioner(id string) (partitioner.Provider, error) {
	if cfg.partitioners != nil {
		if p, ok := cfg.partitioners[id]; ok {
			return p, nil
		}
	}

	return nil, errors.New("partitioner not found: " + id)
}

type partitionerConfig struct {
	Type string `yaml:"type"`

	URL   string `yaml:"url"`
	Token string `yaml:"token"`

	ChunkSize    *int `yaml:"chunkSize"`
	ChunkOverlap *int `yaml:"chunkOverlap"`
}

func (cfg *Config) RegisterPartitioners(f *configFile) error {
	var partitioners []partitioner.Provider

	for id, p := range f.Partitioners {
		partitioner, err := createPartitioner(p)

		if err != nil {
			return err
		}

		partitioners = append(partitioners, partitioner)

		cfg.RegisterPartitioner(p.Type, id, partitioner)
	}

	cfg.RegisterPartitioner("default", "default", multi.New(partitioners...))

	return nil
}

func createPartitioner(cfg partitionerConfig) (partitioner.Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "text":
		return textPartitioner(cfg)

	case "code":
		return codePartitioner(cfg)

	case "azure":
		return azurePartitioner(cfg)

	case "tika":
		return tikaPartitioner(cfg)

	case "unstructured":
		return unstructuredPartitioner(cfg)

	default:
		return nil, errors.New("invalid partitioner type: " + cfg.Type)
	}
}

func textPartitioner(cfg partitionerConfig) (partitioner.Provider, error) {
	var options []text.Option

	if cfg.ChunkSize != nil {
		options = append(options, text.WithChunkSize(*cfg.ChunkSize))
	}

	if cfg.ChunkOverlap != nil {
		options = append(options, text.WithChunkOverlap(*cfg.ChunkOverlap))
	}

	return text.New(options...)
}

func codePartitioner(cfg partitionerConfig) (partitioner.Provider, error) {
	var options []code.Option

	if cfg.ChunkSize != nil {
		options = append(options, code.WithChunkSize(*cfg.ChunkSize))
	}

	if cfg.ChunkOverlap != nil {
		options = append(options, code.WithChunkOverlap(*cfg.ChunkOverlap))
	}

	return code.New(options...)
}

func azurePartitioner(cfg partitionerConfig) (partitioner.Provider, error) {
	var options []azure.Option

	return azure.New(cfg.URL, cfg.Token, options...)
}

func tikaPartitioner(cfg partitionerConfig) (partitioner.Provider, error) {
	var options []tika.Option

	if cfg.ChunkSize != nil {
		options = append(options, tika.WithChunkSize(*cfg.ChunkSize))
	}

	if cfg.ChunkOverlap != nil {
		options = append(options, tika.WithChunkOverlap(*cfg.ChunkOverlap))
	}

	return tika.New(cfg.URL, options...)
}

func unstructuredPartitioner(cfg partitionerConfig) (partitioner.Provider, error) {
	var options []unstructured.Option

	if cfg.URL != "" {
		options = append(options, unstructured.WithURL(cfg.URL))
	}

	if cfg.Token != "" {
		options = append(options, unstructured.WithToken(cfg.Token))
	}

	if cfg.ChunkSize != nil {
		options = append(options, unstructured.WithChunkSize(*cfg.ChunkSize))
	}

	if cfg.ChunkOverlap != nil {
		options = append(options, unstructured.WithChunkOverlap(*cfg.ChunkOverlap))
	}

	return unstructured.New(options...)
}
