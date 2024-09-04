package tika

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/adrianliechti/llama/pkg/partitioner"
	"github.com/adrianliechti/llama/pkg/text"
)

var _ partitioner.Provider = &Client{}

type Client struct {
	client *http.Client

	url string

	chunkSize    int
	chunkOverlap int
}

func New(url string, options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,

		url: url,

		chunkSize:    4000,
		chunkOverlap: 200,
	}

	for _, option := range options {
		option(c)
	}

	if c.url == "" {
		return nil, errors.New("invalid url")
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

	content := text.Normalize(response.Content)

	splitter := text.NewSplitter()
	splitter.ChunkSize = c.chunkSize
	splitter.ChunkOverlap = c.chunkOverlap

	blocks := splitter.Split(content)

	var result []partitioner.Partition

	for i, b := range blocks {
		p := partitioner.Partition{
			ID:      fmt.Sprintf("%s#%d", input.Name, i+1),
			Content: b,
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
