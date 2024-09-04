package azure

import (
	"net/http"
)

type Config struct {
	client *http.Client

	url   string
	token string

	language string
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

func WithLanguage(language string) Option {
	return func(c *Config) {
		c.language = language
	}
}
