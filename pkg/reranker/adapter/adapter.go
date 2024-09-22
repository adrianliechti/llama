package adapter

import (
	"context"
	"math"
	"sort"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/reranker"
)

var _ reranker.Provider = (*Adapter)(nil)

type Adapter struct {
	embedder provider.Embedder
}

func FromEmbedder(embedder provider.Embedder) *Adapter {
	return &Adapter{
		embedder: embedder,
	}
}

func (a *Adapter) Rerank(ctx context.Context, query string, inputs []string, options *reranker.RerankOptions) ([]reranker.Result, error) {
	result, err := a.embedder.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	var results []reranker.Result

	for _, input := range inputs {
		embedding, err := a.embedder.Embed(ctx, input)

		if err != nil {
			return nil, err
		}

		score := cosineSimilarity(result.Data, embedding.Data)

		result := reranker.Result{
			Content: input,
			Score:   float64(score),
		}

		results = append(results, result)
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
