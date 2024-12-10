package replicate

import (
	"net/http"

	"github.com/replicate/replicate-go"
)

type Config struct {
	token string
	model string

	client *http.Client
}

type Option func(*Config)

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.client = client
	}
}

func WithToken(token string) Option {
	return func(c *Config) {
		c.token = token
	}
}

func WithModel(model string) Option {
	return func(c *Config) {
		c.model = model
	}
}

func (c *Config) Options() []replicate.ClientOption {
	options := []replicate.ClientOption{}

	if c.token != "" {
		options = append(options, replicate.WithToken(c.token))
	}

	return options
}
