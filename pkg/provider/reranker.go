package provider

import (
	"context"
	"math"
	"sort"
)

type Reranker interface {
	Rerank(ctx context.Context, query string, inputs []string, options *RerankOptions) ([]Result, error)
}

type RerankOptions struct {
	Limit *int
}

type embedderWrapper struct {
	embedder Embedder
}

func FromEmbedder(embedder Embedder) Reranker {
	return embedderWrapper{
		embedder: embedder,
	}
}

func (e embedderWrapper) Rerank(ctx context.Context, query string, inputs []string, options *RerankOptions) ([]Result, error) {
	result, err := e.embedder.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	var results []Result

	for _, input := range inputs {
		embedding, err := e.embedder.Embed(ctx, input)

		if err != nil {
			return nil, err
		}

		score := cosineSimilarity(result.Data, embedding.Data)

		result := Result{
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
