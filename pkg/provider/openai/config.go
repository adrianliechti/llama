package openai

import (
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Config struct {
	URL string

	Token string
	Model string

	Client *http.Client
}

type Option func(*Config)

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.Client = client
	}
}

func WithURL(url string) Option {
	return func(c *Config) {
		c.URL = url
	}
}

func WithToken(token string) Option {
	return func(c *Config) {
		c.Token = token
	}
}

func WithModel(model string) Option {
	return func(c *Config) {
		c.Model = model
	}
}

func (c *Config) newClient() *openai.Client {
	config := openai.DefaultConfig(c.Token)

	if c.URL != "" {
		config.BaseURL = c.URL
	}

	if c.Client != nil {
		config.HTTPClient = c.Client
	}

	if strings.Contains(c.URL, "openai.azure.com") {
		config = openai.DefaultAzureConfig(c.Token, c.URL)
	}

	return openai.NewClientWithConfig(config)
}
