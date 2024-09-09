package openai

import (
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
	"golang.org/x/time/rate"
)

type Config struct {
	url string

	token string
	model string

	client  *http.Client
	limiter *rate.Limiter
}

type Option func(*Config)

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

func WithClient(client *http.Client) Option {
	return func(c *Config) {
		c.client = client
	}
}

func WithLimiter(limiter *rate.Limiter) Option {
	return func(c *Config) {
		c.limiter = limiter
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
