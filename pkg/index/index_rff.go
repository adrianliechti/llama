package index

import (
	"sort"
)

func ReciprocalRankFusion(k float32, lists ...[]Result) []Result {
	scores := make(map[string]float32)
	documents := make(map[string]Document)

	for _, list := range lists {
		for i, r := range list {
			score := 1.0 / (float32(i) + float32(k))

			scores[r.ID] += score
			documents[r.ID] = r.Document
		}
	}

	var results []Result

	for id, score := range scores {
		results = append(results, Result{
			Score:    score,
			Document: documents[id],
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
