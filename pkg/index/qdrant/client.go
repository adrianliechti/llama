package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/index"
	"github.com/adrianliechti/llama/pkg/to"

	"github.com/google/uuid"
)

var _ index.Provider = &Client{}

type Client struct {
	client *http.Client

	url string

	embedder  index.Embedder
	namespace string
}

func New(url string, namespace string, options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,

		url: url,

		namespace: namespace,
	}

	for _, option := range options {
		option(c)
	}

	if c.embedder == nil {
		return nil, errors.New("embedder is required")
	}

	if c.namespace == "" {
		return nil, errors.New("namespace is required")
	}

	return c, nil
}

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	if err := c.ensureCollection(c.namespace); err != nil {
		return nil, err
	}

	var offset = 0

	var points []point

	for {
		body := map[string]any{
			"offset": offset,

			"with_vector":  true,
			"with_payload": true,
		}

		u, _ := url.JoinPath(c.url, "collections/"+c.namespace+"/points/scroll")

		req, _ := http.NewRequest("POST", u, jsonReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		var result scrollResult

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		points = append(points, result.Result.Points...)

		if offset <= 0 {
			break
		}

		offset = result.Result.NextPageOffset
	}

	var documents []index.Document

	for _, p := range points {
		documents = append(documents, index.Document{
			ID: p.ID,

			Title:    p.Payload.Title,
			Content:  p.Payload.Content,
			Location: p.Payload.Location,

			Metadata: p.Payload.Metadata,

			Embedding: p.Vector,
		})
	}

	return documents, nil
}

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	if len(documents) == 0 {
		return nil
	}

	if err := c.ensureCollection(c.namespace); err != nil {
		return err
	}

	u, _ := url.JoinPath(c.url, "/collections/"+c.namespace+"/points")

	var points []point

	for _, d := range documents {
		if d.ID == "" {
			d.ID = uuid.NewString()
		}

		if len(d.Embedding) == 0 && c.embedder != nil {
			embedding, err := c.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding.Data
		}

		points = append(points, point{
			ID:     convertID(d.ID),
			Vector: d.Embedding,

			Payload: payload{
				Title:    d.Title,
				Content:  d.Content,
				Location: d.Location,

				Metadata: d.Metadata,
			}})

	}

	body := map[string]any{
		"points": points,
	}

	req, _ := http.NewRequest("PUT", u+"?wait=true", jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return convertError(resp)
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}

	if err := c.ensureCollection(c.namespace); err != nil {
		return err
	}

	var points []string

	for _, id := range ids {
		points = append(points, convertID(id))
	}

	u, _ := url.JoinPath(c.url, "collections/"+c.namespace+"/points/delete")

	body := map[string]any{
		"points": points,
	}

	resp, err := c.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = new(index.QueryOptions)
	}

	if options.Limit == nil {
		options.Limit = to.Ptr(10)
	}

	if err := c.ensureCollection(c.namespace); err != nil {
		return nil, err
	}

	embedding, err := c.embedder.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	u, _ := url.JoinPath(c.url, "collections/"+c.namespace+"/points/search")

	body := map[string]any{
		"vector": embedding.Data,
		"limit":  options.Limit,

		"with_vector":  true,
		"with_payload": true,
	}

	req, _ := http.NewRequest("POST", u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var result queryResult

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var results []index.Result

	for _, r := range result.Result {
		results = append(results, index.Result{
			Score: r.Score,

			Document: index.Document{
				ID: r.ID,

				Title:    r.Payload.Title,
				Content:  r.Payload.Content,
				Location: r.Payload.Location,

				Metadata: r.Payload.Metadata,

				Embedding: r.Vector,
			},
		})
	}

	return results, nil
}

func (c *Client) ensureCollection(name string) error {
	u, _ := url.JoinPath(c.url, "/collections/"+name)

	resp, err := c.client.Get(u)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		embeddings, err := c.embedder.Embed(context.Background(), "init")

		if err != nil {
			return err
		}

		body := map[string]any{
			"vectors": map[string]any{
				"size":     len(embeddings.Data),
				"distance": "Cosine",
			},
		}

		req, _ := http.NewRequest("PUT", u, jsonReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.client.Do(req)

		if err != nil {
			return err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("unable to ensure collection")
	}

	return nil
}

func convertID(id string) string {
	if id == "" {
		return uuid.NewString()
	}

	if _, err := uuid.Parse(id); err == nil {
		return id
	}

	return uuid.NewMD5(uuid.NameSpaceOID, []byte(id)).String()
}

func convertError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)
	text := string(data)

	println(text)

	return errors.New(http.StatusText(resp.StatusCode))
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
