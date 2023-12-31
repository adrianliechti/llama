package memory

import (
	"context"
	"errors"
	"math"
	"sort"

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

func (m *Memory) Search(ctx context.Context, embedding []float32, options *index.SearchOptions) ([]index.Result, error) {
	if options == nil {
		options = &index.SearchOptions{}
	}

	results := make([]index.Result, 0)

	for _, d := range m.documents {
		r := index.Result{
			Document: d,

			Distance: 1 - cosineSimilarity(embedding, d.Embedding),
		}

		if options.TopP <= 0 || r.Distance > options.TopP {
			continue
		}

		results = append(results, r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})

	topK := options.TopK

	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

func cosineSimilarity(a []float32, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	count := len(a)

	dotproduct := 0.0

	magnitudeA := 0.0
	magnitudeB := 0.0

	for k := 0; k < count; k++ {
		valA := float64(a[k])
		valB := float64(b[k])

		dotproduct += valA * valB

		magnitudeA += math.Pow(valA, 2)
		magnitudeB += math.Pow(valB, 2)
	}

	magnitudeA = math.Sqrt(magnitudeA)
	magnitudeB = math.Sqrt(magnitudeB)

	cosine := dotproduct / magnitudeA * magnitudeB
	return float32(cosine)
}
