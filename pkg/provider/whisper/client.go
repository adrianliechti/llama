package whisper

import (
	"errors"
	"io"
	"net/http"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Client struct {
	provider.Transcriber
}

func New(url string, options ...Option) (*Client, error) {
	t, err := NewTranscriber(url, options...)

	if err != nil {
		return nil, err
	}

	return &Client{
		Transcriber: t,
	}, nil
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}
