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
	"path"
	"slices"
	"strings"

	"github.com/adrianliechti/llama/pkg/partitioner"
)

var _ partitioner.Provider = &Client{}

type Client struct {
	client *http.Client

	url   string
	token string

	chunkSize     int
	chunkOverlap  int
	chunkStrategy string
}

func New(options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,

		url: "https://api.unstructured.io/general/v0/general",

		chunkSize:     4000,
		chunkOverlap:  500,
		chunkStrategy: "by_title",
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) Partition(ctx context.Context, input partitioner.File, options *partitioner.PartitionOptions) ([]partitioner.Partition, error) {
	if options == nil {
		options = &partitioner.PartitionOptions{}
	}

	if !isSupported(input) {
		return nil, partitioner.ErrUnsupported
	}

	url, _ := url.JoinPath(c.url, "/general/v0/general")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	w.WriteField("strategy", "auto")

	if c.chunkStrategy != "" {
		w.WriteField("chunking_strategy", c.chunkStrategy)
	}

	if c.chunkSize > 0 {
		w.WriteField("max_characters", fmt.Sprintf("%d", c.chunkSize))
	}

	if c.chunkOverlap > 0 {
		w.WriteField("overlap", fmt.Sprintf("%d", c.chunkOverlap))
	}

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

	var result []partitioner.Partition

	for _, e := range elements {
		p := partitioner.Partition{
			ID:      e.ID,
			Content: e.Text,
		}

		result = append(result, p)
	}

	return result, nil
}

func isSupported(input partitioner.File) bool {
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
