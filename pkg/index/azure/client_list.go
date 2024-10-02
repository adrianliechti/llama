package azure

import (
	"context"

	"github.com/adrianliechti/llama/pkg/index"
)

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	results, err := c.Query(ctx, "*", &index.QueryOptions{})

	if err != nil {
		return nil, err
	}

	var result []index.Document

	for _, r := range results {
		result = append(result, r.Document)
	}

	return result, nil
}
