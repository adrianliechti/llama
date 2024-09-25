package flux

import (
	"net/http"
)

const (
	FluxPro     string = "black-forest-labs/flux-pro"
	FluxDev     string = "black-forest-labs/flux-dev"
	FluxSchnell string = "black-forest-labs/flux-schnell"

	FluxDevRealism string = "xlabs-ai/flux-dev-realism"
)

var (
	SupportedModels = []string{
		FluxPro,
		FluxDev,
		FluxSchnell,

		FluxDevRealism,
	}

	ModelVersion = map[string]string{
		FluxDevRealism: "39b3434f194f87a900d1bc2b6d4b983e90f0dde1d5022c27b52c143d670758fa",
	}
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
