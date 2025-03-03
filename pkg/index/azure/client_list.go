package azure

import (
	"context"

	"github.com/adrianliechti/wingman/pkg/index"
)

func (c *Client) List(ctx context.Context, options *index.ListOptions) (*index.Page[index.Document], error) {
	results, err := c.Query(ctx, "*", &index.QueryOptions{})

	if err != nil {
		return nil, err
	}

	var items []index.Document

	for _, r := range results {
		items = append(items, r.Document)
	}

	page := index.Page[index.Document]{
		Items: items,
	}

	return &page, nil
}
