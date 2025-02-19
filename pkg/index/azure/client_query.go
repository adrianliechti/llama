package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/to"
)

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = new(index.QueryOptions)
	}

	if options.Limit == nil {
		options.Limit = to.Ptr(10)
	}

	queries := map[string]string{
		"search": query,
	}

	if options.Limit != nil {
		queries["$top"] = fmt.Sprintf("%d", *options.Limit)
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", c.requestURL("/indexes/"+c.namespace+"/docs", queries), nil)
	req.Header.Set("api-key", c.token)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result Results

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var results []index.Result

	for _, r := range result.Value {
		result := index.Result{
			Document: index.Document{
				ID: r.ID(),

				Title:   r.Title(),
				Source:  r.Source(),
				Content: r.Content(),

				Metadata: r.Metadata(),
			},
		}

		results = append(results, result)
	}

	return results, nil
}
