package unstructured

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/extractor"
)

var _ extractor.Provider = &Client{}

type Client struct {
	url string

	client *http.Client

	chunkSize    int
	chunkOverlap int
}

type Option func(*Client)

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	c := &Client{
		url: url,

		client: http.DefaultClient,

		chunkSize:    4000,
		chunkOverlap: 200,
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

func WithChunkSize(size int) Option {
	return func(c *Client) {
		c.chunkSize = size
	}
}

func WithChunkOverlap(overlap int) Option {
	return func(c *Client) {
		c.chunkOverlap = overlap
	}
}

func (c *Client) Extract(ctx context.Context, input extractor.File, options *extractor.ExtractOptions) (*extractor.Document, error) {
	if options == nil {
		options = &extractor.ExtractOptions{}
	}

	url, _ := url.JoinPath(c.url, "/general/v0/general")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	w.WriteField("strategy", "auto")
	w.WriteField("languages", "eng")
	w.WriteField("languages", "deu")
	w.WriteField("chunking_strategy", "by_title")
	w.WriteField("max_characters", fmt.Sprintf("%d", c.chunkSize))
	w.WriteField("overlap", fmt.Sprintf("%d", c.chunkOverlap))
	w.WriteField("skip_infer_table_types", "")
	w.WriteField("pdf_infer_table_structure", "true")

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

	result := extractor.Document{
		Name: input.Name,
	}

	if len(elements) > 0 {
		result.Name = elements[0].Metadata.FileName
	}

	for _, e := range elements {
		block := extractor.Block{
			ID:      e.ID,
			Content: e.Text,
		}

		result.Blocks = append(result.Blocks, block)
	}

	return &result, nil
}
