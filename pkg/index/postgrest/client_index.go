package postgrest

import (
	"context"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/index"
)

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	body := []Document{}

	for _, d := range documents {
		if len(d.Embedding) == 0 && c.embedder != nil {
			embedding, err := c.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding.Data
		}

		item := Document{
			ID: d.ID,

			Title:    d.Title,
			Content:  d.Content,
			Location: d.Location,

			Embedding: d.Embedding,
		}

		body = append(body, item)
	}

	url, _ := url.JoinPath(c.url, "docs")

	req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates")

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return convertError(resp)
	}

	return nil
}
