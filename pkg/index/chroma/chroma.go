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

var _ index.Index = &Chroma{}

type Chroma struct {
	baesURL *url.URL

	client     *http.Client
	collection *collection
}

func New(addr, name string) (*Chroma, error) {
	c := http.DefaultClient

	u, err := url.Parse(addr)

	if err != nil {
		return nil, err
	}

	chroma := &Chroma{
		baesURL: u,
		client:  c,
	}

	collection, err := chroma.createCollection(name)

	if err != nil {
		return nil, err
	}

	chroma.collection = collection

	return chroma, nil
}

func (c *Chroma) Index(ctx context.Context, documents ...index.Document) error {
	u, _ := url.JoinPath(c.baesURL.String(), "/api/v1/collections/"+c.collection.ID+"/upsert")

	payload := embeddings{
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

		payload.IDs[i] = id

		payload.Embeddings[i] = d.Embeddings

		payload.Documents[i] = d.Content
		payload.Metadatas[i] = d.Metadata
	}

	body, _ := json.Marshal(payload)

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

func (c *Chroma) Search(ctx context.Context, embeddings []float32) ([]index.Result, error) {
	u, _ := url.JoinPath(c.baesURL.String(), "/api/v1/collections/"+c.collection.ID+"/query")

	payload, _ := json.Marshal(map[string]any{
		"query_embeddings": [][]float32{
			embeddings,
		},
	})

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(payload))

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

					// Embeddings: float32to64(result.Embeddings[i]),

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
	u, _ := url.JoinPath(c.baesURL.String(), "/api/v1/collections")

	payload, _ := json.Marshal(map[string]any{
		"name":          name,
		"get_or_create": true,
	})

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(payload))

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
