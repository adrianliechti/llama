package client

import (
	"net/http"
)

type RequestOption = func(*RequestConfig) error

type RequestConfig struct {
	Client *http.Client

	URL   string
	Token string
}

func WithClient(client *http.Client) RequestOption {
	return func(c *RequestConfig) error {
		c.Client = client
		return nil
	}
}

func WithURL(url string) RequestOption {
	return func(c *RequestConfig) error {
		c.URL = url
		return nil
	}
}

func WithToken(token string) RequestOption {
	return func(c *RequestConfig) error {
		c.Token = token
		return nil
	}
}
