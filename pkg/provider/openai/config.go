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

func (c *Config) newClient() *openai.Client {
	config := openai.DefaultConfig(c.token)

	if c.url != "" {
		config.BaseURL = c.url
	}

	if strings.Contains(c.url, "openai.azure.com") {
		config = openai.DefaultAzureConfig(c.token, c.url)
		config.APIVersion = "2024-02-01"
	}

	if c.client != nil {
		config.HTTPClient = c.client
	}

	return openai.NewClientWithConfig(config)
}
