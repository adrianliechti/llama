package postgrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/adrianliechti/wingman/pkg/index"
)

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = new(index.QueryOptions)
	}

	limit := 10

	if options.Limit != nil {
		limit = *options.Limit
	}

	embedding, err := c.embedder.Embed(ctx, []string{query})

	if err != nil {
		return nil, err
	}

	body := map[string]any{
		"query_embedding": embedding.Embeddings[0],
		"limit_count":     limit,
	}

	url, _ := url.JoinPath(c.url, "/rpc/find_similar_docs")

	req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var documents []Document

	if err := json.NewDecoder(resp.Body).Decode(&documents); err != nil {
		return nil, err
	}

	var result []index.Result

	for _, doc := range documents {
		result = append(result, index.Result{
			Document: index.Document{
				ID: doc.ID,

				Title:   doc.Title,
				Source:  doc.Source,
				Content: doc.Content,
			},
		})
	}

	return result, nil
}
