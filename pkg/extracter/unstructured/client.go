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

	"github.com/adrianliechti/llama/pkg/extracter"
	"github.com/adrianliechti/llama/pkg/text"
)

var _ extracter.Provider = &Client{}

type Client struct {
	url string

	client *http.Client
}

type Option func(*Client)

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	c := &Client{
		url: url,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func (c *Client) Extract(ctx context.Context, input extracter.File, options *extracter.ExtractOptions) (*extracter.Document, error) {
	if options == nil {
		options = &extracter.ExtractOptions{}
	}

	url, _ := url.JoinPath(c.url, "/general/v0/general")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	w.WriteField("strategy", "auto")
	w.WriteField("languages", "eng")
	w.WriteField("pdf_infer_table_structure", "true")
	w.WriteField("chunking_strategy", "by_title")

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

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to convert")
	}

	defer resp.Body.Close()

	var elements []Element

	if err := json.NewDecoder(resp.Body).Decode(&elements); err != nil {
		return nil, err
	}

	result := extracter.Document{
		Name: input.Name,
	}

	if len(elements) > 0 {
		result.Name = elements[0].Metadata.FileName
	}

	for _, e := range elements {
		block := extracter.Block{
			ID:      e.ID,
			Content: text.Normalize(e.Text),
		}

		result.Blocks = append(result.Blocks, block)
	}

	return &result, nil
}
