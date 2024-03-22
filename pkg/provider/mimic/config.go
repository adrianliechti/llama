package mimic

import (
	"net/http"
)

type Config struct {
	url string

	client *http.Client
}

type Option func(*Config)

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.client = client
	}
}
