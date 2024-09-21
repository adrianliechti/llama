package unstructured

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
)

var _ extractor.Provider = &Client{}

type Client struct {
	client *http.Client

	url   string
	token string

	strategy Strategy
}

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		url = "https://api.unstructured.io/general/v0/general"
	}

	c := &Client{
		client: http.DefaultClient,

		url: url,

		strategy: "auto",
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) Extract(ctx context.Context, input extractor.File, options *extractor.ExtractOptions) (*extractor.Document, error) {
	if options == nil {
		options = new(extractor.ExtractOptions)
	}

	if !isSupported(input) {
		return nil, extractor.ErrUnsupported
	}

	url, _ := url.JoinPath(c.url, "/general/v0/general")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	w.WriteField("strategy", string(c.strategy))

	file, err := w.CreateFormFile("files", input.Name)

	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(file, input.Content); err != nil {
		return nil, err
	}

	w.Close()

	req, _ := http.NewRequestWithContext(ctx, "POST", url, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var elements []Element

	if err := json.NewDecoder(resp.Body).Decode(&elements); err != nil {
		return nil, err
	}

	name := input.Name

	var builder strings.Builder

	for _, e := range elements {
		builder.WriteString(e.Text)
		builder.WriteString("\n")

		if name == "" {
			name = e.Metadata.FileName
		}
	}

	return &extractor.Document{
		Name:    name,
		Content: builder.String(),
	}, nil
}

func isSupported(input extractor.File) bool {
	if input.Content == nil {
		return false
	}

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
