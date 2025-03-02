package postgrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/adrianliechti/wingman/pkg/index"
)

func (c *Client) List(ctx context.Context, options *index.ListOptions) (*index.Page[index.Document], error) {
	url, _ := url.JoinPath(c.url, "/docs")

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

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

	var items []index.Document

	for _, doc := range documents {
		items = append(items, index.Document{
			ID: doc.ID,

			Title:   doc.Title,
			Source:  doc.Source,
			Content: doc.Content,
		})
	}

	page := index.Page[index.Document]{
		Items: items,
	}

	return &page, nil
}
