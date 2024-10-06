package openai

import (
	"net/http"

	"github.com/openai/openai-go/option"
)

type Config struct {
	url string

	token string
	model string

	client *http.Client
}

type Option func(*Config)

func WithToken(token string) Option {
	return func(c *Config) {
		c.token = token
	}
}

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.client = client
	}
}

func (c *Config) Options() []option.RequestOption {
	options := make([]option.RequestOption, 0)

	options = append(options, option.WithEnvironmentProduction())

	if c.url != "" {
		options = append(options, option.WithBaseURL(c.url))
	}

	if c.token != "" {
		options = append(options, option.WithAPIKey(c.token))
	}

	return options
}
