package tesseract

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
	"strings"

	"github.com/adrianliechti/llama/pkg/extractor"
	"github.com/adrianliechti/llama/pkg/text"
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

	url, _ := url.JoinPath(c.url, "/tesseract")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	w.WriteField("options", c.optionsJSON())

	file, err := w.CreateFormFile("file", input.Name)

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

	var data Result

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Data.Stderr != "" {
		return nil, errors.New(data.Data.Stderr)
	}

	result := extractor.Document{
		Name: input.Name,
	}

	output := data.Data.Stdout
	output = strings.ReplaceAll(output, "\r\n", "\n")

	var lines []string

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimLeft(line, ". ")
		lines = append(lines, line)
	}

	splitter := text.NewSplitter()
	splitter.ChunkSize = c.chunkSize
	splitter.ChunkOverlap = c.chunkOverlap

	chunks := splitter.Split(strings.Join(lines, "\n"))

	for i, chunk := range chunks {
		chunk = strings.ReplaceAll(chunk, "\n\n", "\n")

		block := []extractor.Block{
			{
				ID:      fmt.Sprintf("%s#%d", result.Name, i),
				Content: chunk,
			},
		}

		result.Blocks = append(result.Blocks, block...)
	}

	return &result, nil
}

func (c *Client) optionsJSON() string {
	options := map[string]any{
		"languages": []string{
			"eng",
			"deu",
		},
	}

	value, _ := json.Marshal(options)
	return string(value)
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}
