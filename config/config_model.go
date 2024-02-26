package config

import (
	"github.com/adrianliechti/llama/pkg/provider"
)

func (cfg *Config) RegisterModel(id string) {
	if cfg.models == nil {
		cfg.models = make(map[string]provider.Model)
	}

	if _, ok := cfg.models[id]; ok {
		return
	}

	cfg.models[id] = provider.Model{
		ID: id,
	}
}
