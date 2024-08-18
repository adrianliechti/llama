package flux

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

func NewConfig(options ...Option) *Config {
	c := &Config{
		url: "https://api.replicate.com/",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	return c
}

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
