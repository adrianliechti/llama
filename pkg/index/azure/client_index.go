package azure

import (
	"context"
	"net/http"

	"github.com/adrianliechti/llama/pkg/index"
)

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	if err := c.ensureCollection(ctx, c.namespace); err != nil {
		return err
	}

	items := []map[string]any{}

	for _, d := range documents {
		item := map[string]any{
			"@search.action": "upload",

			"id": d.ID,

			"title":    d.Title,
			"content":  d.Content,
			"location": d.Location,
		}

		if len(d.Metadata) > 0 {
			metadata := []map[string]string{}

			for k, v := range d.Metadata {
				metadata = append(metadata, map[string]string{
					"key":   k,
					"value": v,
				})
			}

			item["metadata"] = metadata
		}

		items = append(items, item)
	}

	body := map[string]any{
		"value": items,
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", c.requestURL("/indexes/"+c.namespace+"/docs/index", nil), jsonReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.token)

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return convertError(resp)
	}

	return nil
}
