package groq

import (
	"net/http"

	"github.com/adrianliechti/wingman/pkg/provider/openai"
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

func WithToken(token string) Option {
	return func(c *Config) {
		c.options = append(c.options, openai.WithToken(token))
	}
}
