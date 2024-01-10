package weaviate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
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

	class string
}

type Option func(*Weaviate)

func New(url, namespace string, options ...Option) (*Weaviate, error) {
	w := &Weaviate{
		url: url,

		client: http.DefaultClient,

		class: namespace,
	}

	for _, option := range options {
		option(w)
	}

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

func (w *Weaviate) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	var vector strings.Builder

	embedding, err := w.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	for i, v := range embedding {
		if i > 0 {
			vector.WriteString(", ")
		}

		vector.WriteString(fmt.Sprintf("%f", v))
	}

	data := executeQueryTemplate(queryData{
		Class: w.class,

		Query:  query,
		Vector: embedding,

		Limit: options.Limit,
		Where: options.Filters,
	})

	body := map[string]any{
		"query": data,
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

		Errors []errorDetail `json:"errors"`
	}

	var result responseType

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Errors) > 0 {
		var errs []error

		for _, e := range result.Errors {
			errs = append(errs, errors.New(e.Message))
		}

		return nil, errors.Join(errs...)
	}

	results := make([]index.Result, 0)

	for _, d := range result.Data.Get[w.class] {
		r := index.Result{
			Document: index.Document{
				ID:      d.Additional.ID,
				Content: d.Content,
			},

			Distance: d.Additional.Distance,
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

func (w *Weaviate) createObject(d index.Document) error {
	properties := maps.Clone(d.Metadata)
	properties["content"] = d.Content

	body := map[string]any{
		"id": d.ID,

		"class":  w.class,
		"vector": d.Embedding,

		"properties": properties,
	}

	u, _ := url.JoinPath(w.url, "/v1/objects")
	resp, err := w.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return convertError(resp)
	}

	return nil
}

func (w *Weaviate) updateObject(d index.Document) error {
	properties := maps.Clone(d.Metadata)
	properties["content"] = d.Content

	body := map[string]any{
		"id": d.ID,

		"class":  w.class,
		"vector": d.Embedding,

		"properties": properties,
	}

	u, _ := url.JoinPath(w.url, "/v1/objects/"+w.class+"/"+d.ID)
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
		return convertError(resp)
	}

	return nil
}

func convertError(resp *http.Response) error {
	type resultType struct {
		Errors []errorDetail `json:"error"`
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

type errorDetail struct {
	Message string `json:"message"`
}

type document struct {
	Content string `json:"content"`

	Additional additional `json:"_additional"`
}

type additional struct {
	ID       string  `json:"id"`
	Distance float32 `json:"distance"`
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
