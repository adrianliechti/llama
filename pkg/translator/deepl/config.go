package deepl

import (
	"net/http"
)

type Option func(*Client)

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

func WithLanguage(language string) Option {
	return func(c *Client) {
		c.language = language
	}
}
