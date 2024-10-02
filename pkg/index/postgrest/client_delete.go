package postgrest

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) Delete(ctx context.Context, ids ...string) error {
	url, _ := url.JoinPath(c.url, "/docs")
	url += "?id=in.(" + strings.Join(ids, ",") + ")"

	req, _ := http.NewRequestWithContext(ctx, "DELETE", url, nil)

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return convertError(resp)
	}

	return nil
}
