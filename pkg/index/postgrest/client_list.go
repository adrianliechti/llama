package postgrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/index"
)

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
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

	var result []index.Document

	for _, doc := range documents {
		result = append(result, index.Document{
			ID: doc.ID,

			Title:    doc.Title,
			Location: doc.Location,

			Content: doc.Content,
		})
	}

	return result, nil
}
