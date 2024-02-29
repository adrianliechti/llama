package deepl

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

var (
	_ provider.Completer  = (*Client)(nil)
	_ provider.Translater = (*Client)(nil)
)

type Client struct {
	url string

	token    string
	language string

	client *http.Client
}

type Option func(*Client)

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		url = "https://api-free.deepl.com"
	}

	c := &Client{
		url: url,

		language: "en",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	if c.url == "" {
		return nil, errors.New("invalid url")
	}

	if c.token == "" {
		return nil, errors.New("invalid token")
	}

	return c, nil
}

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

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
