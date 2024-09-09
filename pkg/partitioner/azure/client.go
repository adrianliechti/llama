package azure

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
	"time"

	"github.com/adrianliechti/llama/pkg/partitioner"
	"github.com/adrianliechti/llama/pkg/text"
)

var _ partitioner.Provider = &Client{}

type Client struct {
	client *http.Client

	url   string
	token string

	chunkSize    int
	chunkOverlap int
}

func New(url, token string, options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,

		url:   url,
		token: token,

		chunkSize:    4000,
		chunkOverlap: 500,
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

	u, _ := url.Parse(strings.TrimRight(c.url, "/") + "/documentintelligence/documentModels/prebuilt-layout:analyze")

	query := u.Query()
	query.Set("api-version", "2024-07-31-preview")

	u.RawQuery = query.Encode()

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), input.Content)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Ocp-Apim-Subscription-Key", c.token)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, convertError(resp)
	}

	operationURL := resp.Header.Get("Operation-Location")

	if operationURL == "" {
		return nil, errors.New("missing operation location")
	}

	var operation AnalyzeOperation

	for {
		req, _ := http.NewRequestWithContext(ctx, "GET", operationURL, nil)
		req.Header.Set("Ocp-Apim-Subscription-Key", c.token)

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, convertError(resp)
		}

		if err := json.NewDecoder(resp.Body).Decode(&operation); err != nil {
			return nil, err
		}

		if operation.Status == OperationStatusRunning || operation.Status == OperationStatusNotStarted {
			time.Sleep(5 * time.Second)
			continue
		}

		if operation.Status != OperationStatusSucceeded {
			return nil, errors.New("operation " + string(operation.Status))
		}

		output, err := convertAnalyzeResult(input, operation.Result, c.chunkSize, c.chunkOverlap)

		if err != nil {
			return nil, err
		}

		return output, nil
	}
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

func convertAnalyzeResult(input partitioner.File, response AnalyzeResult, chunkSize, chunkOverlap int) ([]partitioner.Partition, error) {
	content := text.Normalize(response.Content)

	splitter := text.NewSplitter()
	splitter.ChunkSize = chunkSize
	splitter.ChunkOverlap = chunkOverlap

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
