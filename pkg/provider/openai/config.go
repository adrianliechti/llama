package openai

import (
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
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

func (c *Config) Client() *openai.Client {
	config := openai.DefaultConfig(c.token)

	if c.url != "" {
		config.BaseURL = c.url
	}

	if c.client != nil {
		config.HTTPClient = c.client
	}

	if strings.Contains(c.url, "openai.azure.com") {
		config = openai.DefaultAzureConfig(c.token, c.url)
	}

	return openai.NewClientWithConfig(config)
}
