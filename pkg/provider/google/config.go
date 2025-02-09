package google

import (
	"net/http"

	"google.golang.org/api/option"
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

func (c *Config) Options() []option.ClientOption {
	options := []option.ClientOption{}

	if c.client != nil {
		options = append(options, option.WithHTTPClient(c.client))
	}

	if c.token != "" {
		options = append(options, option.WithAPIKey(c.token))
	}

	return options
}
