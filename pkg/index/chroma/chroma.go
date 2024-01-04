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

var _ index.Provider = &Chroma{}

type Chroma struct {
	url string

	client   *http.Client
	embedder index.Embedder

	namespace string
}

type Option func(*Chroma)

func New(url, namespace string, options ...Option) (*Chroma, error) {
	chroma := &Chroma{
		url: url,

		client: http.DefaultClient,

		namespace: namespace,
	}

	for _, option := range options {
		option(chroma)
	}

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

func (c *Chroma) Embed(ctx context.Context, content string) ([]float32, error) {
	if c.embedder == nil {
		return nil, errors.New("no embedder configured")
	}

	return c.embedder.Embed(ctx, content)
}

func (c *Chroma) Index(ctx context.Context, documents ...index.Document) error {
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
		return errors.New("bad request")
	}

	return nil
}

func (c *Chroma) Search(ctx context.Context, embedding []float32, options *index.SearchOptions) ([]index.Result, error) {
	if options == nil {
		options = &index.SearchOptions{}
	}

	col, err := c.createCollection(c.namespace)

	if err != nil {
		return nil, err
	}

	u, _ := url.JoinPath(c.url, "/api/v1/collections/"+col.ID+"/query")

	body := map[string]any{
		"query_embeddings": [][]float32{
			embedding,
		},

		"include": []string{
			"embeddings",
			"documents",
			"metadatas",
			"distances",
		},
	}

	if options.TopK != nil {
		body["n_results"] = options.TopK
	}

	resp, err := c.client.Post(u, "application/json", jsonReader(body))

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

					Embedding: toFloat32s(result.Embeddings[i][j]),

					Content:  result.Documents[i][j],
					Metadata: result.Metadatas[i][j],
				},
			}

			// if options.TopP != nil && r.Distance > (1-*options.TopP) {
			// 	continue
			// }

			results = append(results, r)
		}
	}

	return results, nil
}

func (c *Chroma) createCollection(name string) (*collection, error) {
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

	Embeddings [][][]float64 `json:"embeddings"`

	Metadatas [][]map[string]any `json:"metadatas"`
	Documents [][]string         `json:"documents"`
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
