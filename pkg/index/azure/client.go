package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/index"
)

var (
	_ index.Provider = (*Client)(nil)
)

type Client struct {
	client *http.Client

	url   string
	token string

	namespace string
}

func New(url, namespace, token string, options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,

		url:   url,
		token: token,

		namespace: namespace,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func (c *Client) requestURL(path string, query map[string]string) string {
	u, _ := url.JoinPath(c.url, path)

	url, _ := url.Parse(u)

	values := url.Query()
	values.Set("api-version", "2024-07-01")

	for k, v := range query {
		values.Set(k, v)
	}

	url.RawQuery = values.Encode()

	return url.String()
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)

	if len(data) == 0 {
		return errors.New(http.StatusText(resp.StatusCode))
	}

	return errors.New(string(data))
}

func (c *Client) ensureCollection(ctx context.Context, name string) error {
	return c.upsertCollection(ctx, name)
}

func (c *Client) upsertCollection(ctx context.Context, name string) error {
	body := map[string]any{
		"name": name,

		"fields": []map[string]any{
			{
				"name": "id",
				"type": "Edm.String",
				"key":  true,
			},
			{
				"name": "title",
				"type": "Edm.String",
			},
			{
				"name": "content",
				"type": "Edm.String",
			},
			{
				"name": "location",
				"type": "Edm.String",
			},
			{
				"name": "metadata",
				"type": "Collection(Edm.ComplexType)",
				"fields": []map[string]any{
					{
						"name": "key",
						"type": "Edm.String",
					},
					{
						"name": "value",
						"type": "Edm.String",
					},
				},
			},
		},
	}

	req, _ := http.NewRequestWithContext(ctx, "PUT", c.requestURL("/indexes/"+name, nil), jsonReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.token)

	resp, err := c.client.Do(req)

	if err != nil {
		return convertError(resp)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return convertError(resp)
	}

	return nil
}
