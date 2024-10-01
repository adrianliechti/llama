package config

import (
	"github.com/adrianliechti/llama/pkg/segmenter"
	"github.com/adrianliechti/llama/pkg/segmenter/text"
)

func (cfg *Config) RegisterSegmenter(id string, p segmenter.Provider) {
	if cfg.segmenter == nil {
		cfg.segmenter = make(map[string]segmenter.Provider)
	}

	cfg.segmenter[id] = p
}

func (cfg *Config) Segmenter(id string) (segmenter.Provider, error) {
	if cfg.segmenter != nil {
		if p, ok := cfg.segmenter[id]; ok {
			return p, nil
		}
	}

	//return nil, errors.New("segmenter not found: " + id)

	return text.New()
}

func (cfg *Config) RegisterSegmenters(f *configFile) error {
	return nil
}
