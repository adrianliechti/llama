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

var _ index.Provider = &Provider{}

type Provider struct {
	embedder index.Embedder
	reranker index.Reranker

	documents map[string]index.Document
}

func New(options ...Option) (*Provider, error) {
	p := &Provider{
		documents: make(map[string]index.Document),
	}

	for _, option := range options {
		option(p)
	}

	if p.embedder == nil {
		return nil, errors.New("embedder is required")
	}

	return p, nil
}

func (p *Provider) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	result := make([]index.Document, 0, len(p.documents))

	for _, d := range p.documents {
		result = append(result, d)
	}

	return result, nil
}

func (p *Provider) Index(ctx context.Context, documents ...index.Document) error {
	for _, d := range documents {
		if d.ID == "" {
			d.ID = uuid.NewString()
		}

		if len(d.Embedding) == 0 && p.embedder != nil {
			embedding, err := p.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding.Data
		}

		if len(d.Embedding) == 0 {
			continue
		}

		p.documents[d.ID] = d
	}

	return nil
}

func (p *Provider) Delete(ctx context.Context, ids ...string) error {
	for _, id := range ids {
		delete(p.documents, id)
	}

	return nil
}

func (p *Provider) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = &index.QueryOptions{}
	}

	if p.embedder == nil {
		return nil, errors.New("no embedder configured")
	}

	embedding, err := p.embedder.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	results := make([]index.Result, 0)

DOCUMENTS:
	for _, d := range p.documents {
		score := cosineSimilarity(embedding.Data, d.Embedding)

		r := index.Result{
			Score:    score,
			Document: d,
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
		return results[i].Score > results[j].Score
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
