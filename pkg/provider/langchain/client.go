package langchain

import (
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Client struct {
	provider.Completer
}

func New(url string, options ...Option) (*Client, error) {
	c, err := NewCompleter(url, options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Completer: c,
	}, nil
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}
