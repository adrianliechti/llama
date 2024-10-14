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

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	result, err := c.client.List(ctx, &ListRequest{})

	if err != nil {
		return nil, err
	}

	return convertDocuments(result.Documents), nil
}

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	_, err := c.client.Index(ctx, &IndexRequest{
		Documents: toDocuments(documents),
	})

	return err
}

func (c *Client) Delete(ctx context.Context, ids ...string) error {
	_, err := c.client.Delete(ctx, &DeleteRequest{
		Ids: ids,
	})

	return err
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = new(index.QueryOptions)
	}

	var limit *int32

	if options.Limit != nil {
		val := int32(*options.Limit)
		limit = &val
	}

	data, err := c.client.Query(ctx, &QueryRequest{
		Query: query,

		Limit: limit,
	})

	if err != nil {
		return nil, err
	}

	return convertResults(data.Results), nil
}

func convertResult(r *Result) index.Result {
	return index.Result{
		Score:    r.Score,
		Document: convertDocument(r.Document),
	}
}

func convertResults(s []*Result) []index.Result {
	var result []index.Result

	for _, r := range s {
		result = append(result, convertResult(r))
	}

	return result
}

func convertDocument(d *Document) index.Document {
	return index.Document{
		ID: d.Id,

		Title:    d.Title,
		Content:  d.Content,
		Location: d.Location,

		Metadata: d.Metadata,

		Embedding: d.Embedding,
	}
}

func convertDocuments(s []*Document) []index.Document {
	var result []index.Document

	for _, d := range s {
		result = append(result, convertDocument(d))
	}

	return result
}

func toDocument(d index.Document) *Document {
	return &Document{
		Id: d.ID,

		Title:    d.Title,
		Content:  d.Content,
		Location: d.Location,

		Metadata: d.Metadata,

		Embedding: d.Embedding,
	}
}

func toDocuments(s []index.Document) []*Document {
	var result []*Document

	for _, d := range s {
		document := toDocument(d)
		result = append(result, document)
	}

	return result
}
