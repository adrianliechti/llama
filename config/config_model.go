package config

import (
	"github.com/adrianliechti/llama/pkg/provider"
)

func (c *Config) RegisterModel(id string) {
	if _, ok := c.models[id]; ok {
		return
	}

	c.models[id] = provider.Model{
		ID: id,
	}
}
