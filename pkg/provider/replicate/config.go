package replicate

import "net/http"

type Config struct {
	url string

	token string
	model string

	stops    string
	template string

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

func WithStops(stops string) Option {
	return func(c *Config) {
		c.stops = stops
	}
}

func WithTemplate(template string) Option {
	return func(c *Config) {
		c.template = template
	}
}
