package memory

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"

	"github.com/google/uuid"
)

var _ index.Provider = &Client{}

type Client struct {
	embedder index.Embedder

	documents map[string]index.Document
}

type Option func(*Client)

func New(options ...Option) (*Client, error) {
	c := &Client{
		documents: make(map[string]index.Document),
	}

	for _, option := range options {
		option(c)
	}

	if c.embedder == nil {
		return nil, errors.New("embedder is required")
	}

	return c, nil
}

func WithEmbedder(embedder index.Embedder) Option {
	return func(c *Client) {
		c.embedder = embedder
	}
}

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	result := make([]index.Document, 0, len(c.documents))

	for _, d := range c.documents {
		result = append(result, d)
	}

	return result, nil
}

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	for _, d := range documents {
		if d.ID == "" {
			d.ID = uuid.NewString()
		}

		if len(d.Embedding) == 0 && c.embedder != nil {
			embedding, err := c.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding
		}

		if len(d.Embedding) == 0 {
			continue
		}

		c.documents[d.ID] = d
	}

	return nil
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = &index.QueryOptions{}
	}

	if c.embedder == nil {
		return nil, errors.New("no embedder configured")
	}

	embedding, err := c.embedder.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	results := make([]index.Result, 0)

DOCUMENTS:
	for _, d := range c.documents {
		r := index.Result{
			Document: d,

			Distance: 1.0 - cosineSimilarity(embedding, d.Embedding),
		}

		if options.Distance != nil {
			if r.Distance > *options.Distance {
				continue
			}
		}

		for k, v := range options.Filters {
			val, ok := d.Metadata[k]

			if !ok {
				continue DOCUMENTS
			}

			if !strings.EqualFold(v, val) {
				continue DOCUMENTS
			}
		}

		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})

	if options.Limit != nil {
		limit := *options.Limit

		if limit > len(results) {
			limit = len(results)
		}

		results = results[:limit]
	}

	return results, nil
}

func cosineSimilarity(a []float32, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	dotproduct := 0.0

	magnitudeA := 0.0
	magnitudeB := 0.0

	for k := 0; k < len(a); k++ {
		valA := float64(a[k])
		valB := float64(b[k])

		dotproduct += valA * valB

		magnitudeA += math.Pow(valA, 2)
		magnitudeB += math.Pow(valB, 2)
	}

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0.0
	}

	return float32(dotproduct / (math.Sqrt(magnitudeA) * math.Sqrt(magnitudeB)))
}
