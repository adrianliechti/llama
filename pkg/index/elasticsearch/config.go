package elasticsearch

import (
	"net/http"
)

type Config struct {
	client *http.Client

	url string

	namespace string
}

type Option func(*Config)

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.client = client
	}
}
