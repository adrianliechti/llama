package weaviate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"

	"github.com/google/uuid"
)

var _ index.Provider = &Weaviate{}

type Weaviate struct {
	url string

	client   *http.Client
	embedder index.Embedder

	class *class
}

type Option func(*Weaviate)

func New(url, namespace string, options ...Option) (*Weaviate, error) {
	w := &Weaviate{
		url: url,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(w)
	}

	class, err := w.getClass(namespace)

	if err != nil {
		return nil, err
	}

	w.class = class

	return w, nil
}

func WithClient(client *http.Client) Option {
	return func(w *Weaviate) {
		w.client = client
	}
}

func WithEmbedder(embedder index.Embedder) Option {
	return func(w *Weaviate) {
		w.embedder = embedder
	}
}

func (w *Weaviate) Embed(ctx context.Context, content string) ([]float32, error) {
	if w.embedder == nil {
		return nil, errors.New("no embedder configured")
	}

	return w.embedder.Embed(ctx, content)
}

func (w *Weaviate) Index(ctx context.Context, documents ...index.Document) error {
	for _, d := range documents {
		d.ID = generateID(d)

		if len(d.Embedding) == 0 && w.embedder != nil {
			embedding, err := w.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding
		}

		err := w.createObject(d)

		if err != nil {
			err = w.updateObject(d)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Weaviate) Query(ctx context.Context, embedding []float32, options *index.QueryOptions) ([]index.Result, error) {
	var vector strings.Builder

	for i, v := range embedding {
		if i > 0 {
			vector.WriteString(", ")
		}

		vector.WriteString(fmt.Sprintf("%f", v))
	}

	limit := 10
	distance := float32(1)

	if options.Limit != nil {
		limit = *options.Limit
	}

	if options.Distance != nil {
		distance = *options.Distance
	}

	query := `{
		Get {
			` + w.class.Class + ` (
				limit: ` + fmt.Sprintf("%d", limit) + `
				nearVector: {
					vector: [` + vector.String() + `]
					distance: ` + fmt.Sprintf("%f", distance) + `
				}
			) {
				content
				_additional {
					distance
				}
			}
		}
	}`

	body := map[string]any{
		"query": query,
	}

	u, _ := url.JoinPath(w.url, "/v1/graphql")
	resp, err := w.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad request")
	}

	type responseType struct {
		Data struct {
			Get map[string][]document `json:"Get"`
		} `json:"data"`
	}

	var result responseType

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	results := make([]index.Result, 0)

	for _, d := range result.Data.Get[w.class.Class] {
		r := index.Result{
			Document: index.Document{
				Content: d.Content,
			},

			Distance: 1 - d.Additional.Certainty,
		}

		results = append(results, r)
	}

	return results, nil
}

func generateID(d index.Document) string {
	if d.ID == "" {
		return uuid.NewString()
	}

	return uuid.NewMD5(uuid.NameSpaceOID, []byte(d.ID)).String()
}

func (w *Weaviate) getClass(name string) (*class, error) {
	u, _ := url.JoinPath(w.url, "/v1/schema/"+name)

	resp, err := w.client.Get(u)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return w.createClass(name)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad request")
	}

	var result class

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (w *Weaviate) createClass(name string) (*class, error) {
	u, _ := url.JoinPath(w.url, "/v1/schema")

	body := map[string]any{
		"class": name,

		"properties": []map[string]any{
			{
				"name": "content",

				"dataType": []string{
					"text",
				},
			},
		},
	}

	resp, err := w.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("bad request")
	}

	var result class

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (w *Weaviate) createObject(d index.Document) error {
	body := map[string]any{
		"id": d.ID,

		"class":  w.class.Class,
		"vector": d.Embedding,

		"properties": map[string]any{
			"content": d.Content,
		},
	}

	u, _ := url.JoinPath(w.url, "/v1/objects")
	resp, err := w.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad request")
	}

	return nil
}

func (w *Weaviate) updateObject(d index.Document) error {
	body := map[string]any{
		"id": d.ID,

		"class":  w.class.Class,
		"vector": d.Embedding,

		"properties": map[string]any{
			"content": d.Content,
		},
	}

	u, _ := url.JoinPath(w.url, "/v1/objects/"+w.class.Class+"/"+d.ID)
	req, err := http.NewRequest(http.MethodPut, u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

	resp, err := w.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad request")
	}

	return nil
}

type class struct {
	Class string `json:"class"`
}

type document struct {
	Content string `json:"content"`

	Additional additional `json:"_additional"`
}

type additional struct {
	Certainty float32 `json:"certainty"`
	Distance  float32 `json:"distance"`
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
