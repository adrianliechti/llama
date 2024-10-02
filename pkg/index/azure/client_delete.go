package azure

import (
	"context"
	"net/http"
)

func (c *Client) Delete(ctx context.Context, ids ...string) error {
	if err := c.ensureCollection(ctx, c.namespace); err != nil {
		return err
	}

	items := []map[string]any{}

	for _, id := range ids {
		item := map[string]any{
			"@search.action": "delete",

			"id": id,
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
