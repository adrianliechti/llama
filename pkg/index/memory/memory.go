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

var _ index.Provider = &Memory{}

type Memory struct {
	embedder index.Embedder

	documents map[string]index.Document
}

type Option func(*Memory)

func New(options ...Option) (*Memory, error) {
	m := &Memory{
		documents: make(map[string]index.Document),
	}

	for _, option := range options {
		option(m)
	}

	if m.embedder == nil {
		return nil, errors.New("embedder is required")
	}

	return m, nil
}

func WithEmbedder(embedder index.Embedder) Option {
	return func(m *Memory) {
		m.embedder = embedder
	}
}

func (m *Memory) Embed(ctx context.Context, content string) ([]float32, error) {
	if m.embedder == nil {
		return nil, errors.New("no embedder configured")
	}

	return m.embedder.Embed(ctx, content)
}

func (m *Memory) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	result := make([]index.Document, 0, len(m.documents))

	for _, d := range m.documents {
		result = append(result, d)
	}

	return result, nil
}

func (m *Memory) Index(ctx context.Context, documents ...index.Document) error {
	for _, d := range documents {
		if d.ID == "" {
			d.ID = uuid.NewString()
		}

		if len(d.Embedding) == 0 && m.embedder != nil {
			embedding, err := m.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding
		}

		m.documents[d.ID] = d
	}

	return nil
}

func (m *Memory) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = &index.QueryOptions{}
	}

	embedding, err := m.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	results := make([]index.Result, 0)

DOCUMENTS:
	for _, d := range m.documents {
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
