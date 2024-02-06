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

var _ index.Provider = &Elasticsearch{}

type Elasticsearch struct {
	url string

	client *http.Client

	namespace string
}

type Option func(*Elasticsearch)

func New(url, namespace string, options ...Option) (*Elasticsearch, error) {
	elasticsearch := &Elasticsearch{
		url: url,

		client: http.DefaultClient,

		namespace: namespace,
	}

	for _, option := range options {
		option(elasticsearch)
	}

	return elasticsearch, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Elasticsearch) {
		c.client = client
	}
}

func (e *Elasticsearch) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	u, _ := url.JoinPath(e.url, "/"+e.namespace+"/_search")

	body := map[string]any{
		"size": 10000,
		"from": 0,
		"query": map[string]any{
			"match_all": map[string]any{},
		},
	}

	req, _ := http.NewRequest(http.MethodGet, u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)

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

func (e *Elasticsearch) Index(ctx context.Context, documents ...index.Document) error {
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

		u, _ := url.JoinPath(e.url, "/"+e.namespace+"/_doc/"+d.ID)
		resp, err := e.client.Post(u, "application/json", jsonReader(body))

		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return convertError(resp)
		}
	}

	return nil
}

func (e *Elasticsearch) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	u, _ := url.JoinPath(e.url, "/"+e.namespace+"/_search")

	body := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":    query,
				"fields":   []string{"content", "metadata.*"},
				"analyzer": "english",
			},
		},
	}

	req, _ := http.NewRequest(http.MethodGet, u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)

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

type Document struct {
	ID string `json:"id"`

	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

type SearchResult struct {
	Hits SearchHits `json:"hits"`
}

type SearchHits struct {
	Hits []SearchHit `json:"hits"`
}

type SearchHit struct {
	Score    float32  `json:"_score"`
	Document Document `json:"_source"`
}
