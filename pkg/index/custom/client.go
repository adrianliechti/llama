package custom

import (
	"context"
	"errors"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	_ index.Provider = (*Client)(nil)
)

type Client struct {
	url string

	client IndexClient
}

type Option func(*Client)

func New(url string, options ...Option) (*Client, error) {
	if url == "" || !strings.HasPrefix(url, "grpc://") {
		return nil, errors.New("invalid url")
	}

	c := &Client{
		url: url,
	}

	for _, option := range options {
		option(c)
	}

	url = strings.TrimPrefix(c.url, "grpc://")

	conn, err := grpc.Dial(url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	c.client = NewIndexClient(conn)

	return c, nil
}

func (*Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	return nil, errors.ErrUnsupported
}

func (*Client) Index(ctx context.Context, documents ...index.Document) error {
	return errors.ErrUnsupported
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = new(index.QueryOptions)
	}

	var limit *int32
	var distance *float32

	if options.Limit != nil {
		val := int32(*options.Limit)
		limit = &val
	}

	if options.Distance != nil {
		val := float32(*options.Distance)
		distance = &val
	}

	data, err := c.client.Query(ctx, &QueryRequest{
		Query: query,

		Limit:    limit,
		Distance: distance,
	})

	if err != nil {
		return nil, err
	}

	var results []index.Result

	for _, r := range data.Results {
		result := index.Result{
			Document: index.Document{
				ID: r.Document.Id,

				Title:    r.Document.Title,
				Content:  r.Document.Content,
				Location: r.Document.Location,

				Metadata: r.Document.Metadata,

				Embedding: r.Document.Embedding,
			},

			Distance: r.Distance,
		}

		results = append(results, result)
	}

	return results, nil
}
