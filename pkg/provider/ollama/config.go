package ollama

import (
	"net/http"

	"github.com/adrianliechti/llama/pkg/template"
)

type Config struct {
	url string

	model    string
	template template.Template

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

func WithTemplate(template template.Template) Option {
	return func(c *Config) {
		c.template = template
	}
}
