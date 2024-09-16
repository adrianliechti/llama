package tika

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/adrianliechti/llama/pkg/converter"
	"github.com/adrianliechti/llama/pkg/text"
)

var _ converter.Provider = &Client{}

type Client struct {
	client *http.Client

	url string
}

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	c := &Client{
		client: http.DefaultClient,

		url: url,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) Convert(ctx context.Context, input converter.File, options *converter.ConvertOptions) (*converter.Document, error) {
	if options == nil {
		options = new(converter.ConvertOptions)
	}

	if !isSupported(input) {
		return nil, converter.ErrUnsupported
	}

	url, _ := url.JoinPath(c.url, "/tika/text")
	req, _ := http.NewRequestWithContext(ctx, "PUT", url, input.Content)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var response TikaResponse

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &converter.Document{
		Content: text.Normalize(response.Content),
	}, nil
}

func isSupported(input converter.File) bool {
	ext := strings.ToLower(path.Ext(input.Name))
	return slices.Contains(SupportedExtensions, ext)
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}
