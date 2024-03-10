package chroma

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

	client   *http.Client
	embedder index.Embedder

	namespace string
}

type Option func(*Client)

func New(url, namespace string, options ...Option) (*Client, error) {
	chroma := &Client{
		url: url,

		client: http.DefaultClient,

		namespace: namespace,
	}

	for _, option := range options {
		option(chroma)
	}

	if chroma.embedder == nil {
		return nil, errors.New("embedder is required")
	}

	return chroma, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func WithEmbedder(embedder index.Embedder) Option {
	return func(c *Client) {
		c.embedder = embedder
	}
}

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	col, err := c.createCollection(c.namespace)

	if err != nil {
		return nil, err
	}

	u, _ := url.JoinPath(c.url, "/api/v1/collections/"+col.ID+"/get")

	body := map[string]any{}

	resp, err := c.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var result getResult

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	results := make([]index.Document, 0)

	for i := range result.IDs {
		r := index.Document{
			ID: result.IDs[i],

			Content:  result.Documents[i],
			Metadata: result.Metadatas[i],
		}

		results = append(results, r)
	}

	return results, nil
}

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	if len(documents) == 0 {
		return nil
	}

	col, err := c.createCollection(c.namespace)

	if err != nil {
		return err
	}

	u, _ := url.JoinPath(c.url, "/api/v1/collections/"+col.ID+"/upsert")

	body := embeddings{
		IDs: make([]string, len(documents)),

		Embeddings: make([][]float32, len(documents)),

		Documents: make([]string, len(documents)),
		Metadatas: make([]map[string]string, len(documents)),
	}

	for i, d := range documents {
		id := d.ID

		if id == "" {
			id = uuid.NewString()
		}

		if len(d.Embedding) == 0 && c.embedder != nil {
			embedding, err := c.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding
		}

		body.IDs[i] = id

		body.Embeddings[i] = d.Embedding

		body.Documents[i] = d.Content
		body.Metadatas[i] = d.Metadata
	}

	resp, err := c.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return convertError(resp)
	}

	return nil
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	if options == nil {
		options = &index.QueryOptions{}
	}

	col, err := c.createCollection(c.namespace)

	if err != nil {
		return nil, err
	}

	embedding, err := c.embedder.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	u, _ := url.JoinPath(c.url, "/api/v1/collections/"+col.ID+"/query")

	body := map[string]any{
		"query_embeddings": [][]float32{
			embedding,
		},

		"include": []string{
			"documents",
			"metadatas",
			"distances",
		},
	}

	if len(options.Filters) > 0 {
		body["where"] = options.Filters
	}

	if options.Limit != nil {
		body["n_results"] = *options.Limit
	}

	resp, err := c.client.Post(u, "application/json", jsonReader(body))

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

	results := make([]index.Result, 0)

	for i := range result.IDs {
		for j := range result.IDs[i] {
			id := result.IDs[i][j]

			content := result.Documents[i][j]
			metadata := result.Metadatas[i][j]

			distance := result.Distances[i][j]

			filename := metadata["filename"]
			filepart := metadata["filepart"]

			title := id
			location := id

			if filename != "" {
				title = filename
				location = filename
			}

			if filepart != "" {
				if location != "" {
					location += "#" + filepart
				}
			}

			r := index.Result{
				Distance: distance,

				Document: index.Document{
					ID: id,

					Title:    title,
					Content:  content,
					Location: location,

					Metadata: metadata,
				},
			}

			if options.Distance != nil {
				if r.Distance > *options.Distance {
					continue
				}
			}

			results = append(results, r)
		}
	}

	return results, nil
}

func (c *Client) createCollection(name string) (*collection, error) {
	u, _ := url.JoinPath(c.url, "/api/v1/collections")

	body := map[string]any{
		"name":          name,
		"get_or_create": true,

		"metadata": map[string]any{
			"hnsw:space": "cosine",
		},
	}

	resp, err := c.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result collection

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func convertError(resp *http.Response) error {
	type resultType struct {
		Errors []errorDetail `json:"detail"`
	}

	var result resultType

	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		var errs []error

		for _, e := range result.Errors {
			errs = append(errs, errors.New(e.Message))
		}

		return errors.Join(errs...)
	}

	return errors.New(http.StatusText(resp.StatusCode))
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}

func toFloat32s(v []float64) []float32 {
	result := make([]float32, len(v))

	for i, x := range v {
		result[i] = float32(x)
	}

	return result
}
