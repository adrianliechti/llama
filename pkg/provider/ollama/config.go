package ollama

import (
	"net/http"
)

type Config struct {
	url string

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

func WithURL(url string) Option {
	return func(c *Config) {
		c.url = url
	}
}

func WithModel(model string) Option {
	return func(c *Config) {
		c.model = model
	}
}
