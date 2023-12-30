package chroma

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/index"

	"github.com/google/uuid"
)

var _ index.Provider = &Chroma{}

type Chroma struct {
	url string

	client *http.Client

	collection *collection

	embedder index.Embedder
}

type Option func(*Chroma)

func New(url, collection string, options ...Option) (*Chroma, error) {
	chroma := &Chroma{
		url: url,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(chroma)
	}

	col, err := chroma.createCollection(collection)

	if err != nil {
		return nil, err
	}

	chroma.collection = col

	return chroma, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Chroma) {
		c.client = client
	}
}

func WithEmbedder(embedder index.Embedder) Option {
	return func(c *Chroma) {
		c.embedder = embedder
	}
}

func (c *Chroma) Index(ctx context.Context, documents ...index.Document) error {
	u, _ := url.JoinPath(c.url, "/api/v1/collections/"+c.collection.ID+"/upsert")

	request := embeddings{
		IDs: make([]string, len(documents)),

		Embeddings: make([][]float32, len(documents)),

		Documents: make([]string, len(documents)),
		Metadatas: make([]map[string]any, len(documents)),
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

		request.IDs[i] = id

		request.Embeddings[i] = d.Embedding

		request.Documents[i] = d.Content
		request.Metadatas[i] = d.Metadata
	}

	body, _ := json.Marshal(request)

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))

	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad request")
	}

	return nil
}

func (c *Chroma) Search(ctx context.Context, embedding []float32, options *index.SearchOptions) ([]index.Result, error) {
	if options == nil {
		options = &index.SearchOptions{}
	}

	u, _ := url.JoinPath(c.url, "/api/v1/collections/"+c.collection.ID+"/query")

	request := map[string]any{
		"query_embeddings": [][]float32{
			embedding,
		},
	}

	if options.Top > 0 {
		request["n_results"] = options.Top
	}

	body, _ := json.Marshal(request)

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad request")
	}

	var result result

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	results := make([]index.Result, 0)

	for i := range result.IDs {
		for j := range result.IDs[i] {
			r := index.Result{
				Distance: result.Distances[i][j],

				Document: index.Document{
					ID: result.IDs[i][j],

					// Embedding: float32to64(result.Embeddings[i]),

					Content:  result.Documents[i][j],
					Metadata: result.Metadatas[i][j],
				},
			}

			results = append(results, r)
		}
	}

	return results, nil
}

func (c *Chroma) createCollection(name string) (*collection, error) {
	u, _ := url.JoinPath(c.url, "/api/v1/collections")

	request := map[string]any{
		"name":          name,
		"get_or_create": true,
	}

	body, _ := json.Marshal(request)

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)

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

type collection struct {
	ID string `json:"id,omitempty"`

	Tenant   string `json:"tenant,omitempty"`
	Database string `json:"database,omitempty"`

	Name     string         `json:"name,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type embeddings struct {
	IDs []string `json:"ids"`

	Embeddings [][]float32 `json:"embeddings"`

	Metadatas []map[string]any `json:"metadatas"`
	Documents []string         `json:"documents"`
}

type result struct {
	IDs [][]string `json:"ids"`

	Distances [][]float32 `json:"distances,omitempty"`

	//Embeddings [][][]float32 `json:"embeddings"`

	Metadatas [][]map[string]any `json:"metadatas"`
	Documents [][]string         `json:"documents"`
}
