package tavily

import (
	"net/http"
)

type Option func(*Tool)

func WithClient(client *http.Client) Option {
	return func(t *Tool) {
		t.client = client
	}
}
