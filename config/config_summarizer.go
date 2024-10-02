package config

import (
	"errors"

	"github.com/adrianliechti/llama/pkg/summarizer"
)

func (cfg *Config) RegisterSummarizer(id string, p summarizer.Provider) {
	if cfg.summarizer == nil {
		cfg.summarizer = make(map[string]summarizer.Provider)
	}

	cfg.summarizer[id] = p
}

func (cfg *Config) Summarizer(id string) (summarizer.Provider, error) {
	if cfg.summarizer != nil {
		if p, ok := cfg.summarizer[id]; ok {
			return p, nil
		}
	}

	return nil, errors.New("summarizer not found: " + id)
}

func (cfg *Config) RegisterSummarizers(f *configFile) error {
	return nil
}
