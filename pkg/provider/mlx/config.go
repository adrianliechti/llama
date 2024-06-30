package mlx

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider/openai"
)

type Config struct {
	options []openai.Option
}

type Option func(*Config)

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.options = append(c.options, openai.WithClient(client))
	}
}

func WithModel(model string) Option {
	return func(c *Config) {
		c.options = append(c.options, openai.WithModel(model))
	}
}
