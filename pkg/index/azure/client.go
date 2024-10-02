package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/to"
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

	u, _ := url.JoinPath(c.url, "/indexes/"+c.namespace+"/docs/index")
	url, _ := url.Parse(u)

	values := url.Query()
	values.Set("api-version", "2024-07-01")

	url.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "POST", url.String(), jsonReader(body))
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

	u, _ := url.JoinPath(c.url, "/indexes/"+c.namespace+"/docs/index")
	url, _ := url.Parse(u)

	values := url.Query()
	values.Set("api-version", "2024-07-01")

	url.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "POST", url.String(), jsonReader(body))
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

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = new(index.QueryOptions)
	}

	if options.Limit == nil {
		options.Limit = to.Ptr(10)
	}

	u, _ := url.JoinPath(c.url, "/indexes/"+c.namespace+"/docs")
	url, _ := url.Parse(u)

	values := url.Query()
	values.Set("search", query)

	if options.Limit != nil {
		values.Set("$top", fmt.Sprintf("%d", *options.Limit))
	}

	values.Set("api-version", "2024-07-01")

	url.RawQuery = values.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
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

				Title:    r.Title(),
				Content:  r.Content(),
				Location: r.Location(),

				Metadata: r.Metadata(),
			},
		}

		results = append(results, result)
	}

	return results, nil
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
	u, _ := url.JoinPath(c.url, "/indexes/"+name)
	url, _ := url.Parse(u)

	values := url.Query()
	values.Set("api-version", "2024-07-01")

	url.RawQuery = values.Encode()

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

	req, _ := http.NewRequestWithContext(ctx, "PUT", url.String(), jsonReader(body))
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
