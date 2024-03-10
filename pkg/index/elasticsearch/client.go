package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/index"

	"github.com/google/uuid"
)

var _ index.Provider = &Client{}

type Client struct {
	url string

	client *http.Client

	namespace string
}

type Option func(*Client)

func New(url, namespace string, options ...Option) (*Client, error) {
	c := &Client{
		url: url,

		client: http.DefaultClient,

		namespace: namespace,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	u, _ := url.JoinPath(c.url, "/"+c.namespace+"/_search")

	body := map[string]any{
		"size": 10000,
		"from": 0,
		"query": map[string]any{
			"match_all": map[string]any{},
		},
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var result SearchResult

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var results []index.Document

	for _, hit := range result.Hits.Hits {
		results = append(results, index.Document{
			ID:       hit.Document.ID,
			Content:  hit.Document.Content,
			Metadata: hit.Document.Metadata,
		})
	}

	return results, nil
}

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	if len(documents) == 0 {
		return nil
	}

	for _, d := range documents {
		d.ID = generateID(d)

		body := Document{
			ID: d.ID,

			Content:  d.Content,
			Metadata: d.Metadata,
		}

		u, _ := url.JoinPath(c.url, "/"+c.namespace+"/_doc/"+d.ID)
		resp, err := c.client.Post(u, "application/json", jsonReader(body))

		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return convertError(resp)
		}
	}

	return nil
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	u, _ := url.JoinPath(c.url, "/"+c.namespace+"/_search")

	body := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":    query,
				"fields":   []string{"content", "metadata.*"},
				"analyzer": "english",
			},
		},
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var result SearchResult

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var results []index.Result

	for _, hit := range result.Hits.Hits {
		results = append(results, index.Result{
			Document: index.Document{
				ID:       hit.Document.ID,
				Content:  hit.Document.Content,
				Metadata: hit.Document.Metadata,
			},
		})
	}

	return results, nil
}

func generateID(d index.Document) string {
	if d.ID == "" {
		return uuid.NewString()
	}

	return uuid.NewMD5(uuid.NameSpaceOID, []byte(d.ID)).String()
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}

func convertError(resp *http.Response) error {
	type resultType struct {
		Error string `json:"error"`
	}

	var result resultType

	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		return errors.New(result.Error)
	}

	return errors.New(http.StatusText(resp.StatusCode))
}
