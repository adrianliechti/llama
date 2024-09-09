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
	"slices"
	"strings"

	"github.com/adrianliechti/llama/pkg/index"

	"github.com/google/uuid"
)

var _ index.Provider = &Client{}

type Client struct {
	client *http.Client

	url string

	embedder index.Embedder
	class    string
}

func New(url, namespace string, options ...Option) (*Client, error) {
	c := &Client{
		client: http.DefaultClient,

		url: url,

		class: namespace,
	}

	for _, option := range options {
		option(c)
	}

	if c.embedder == nil {
		return nil, errors.New("embedder is required")
	}

	if c.class == "" {
		return nil, errors.New("namespace is required")
	}

	return c, nil
}

func (c *Client) List(ctx context.Context, options *index.ListOptions) ([]index.Document, error) {
	if options == nil {
		options = new(index.ListOptions)
	}

	limit := 50
	cursor := ""

	results := make([]index.Document, 0)

	type pageType struct {
		Objects []Object `json:"objects"`
	}

	for {
		query := url.Values{}
		query.Set("class", c.class)
		query.Set("limit", fmt.Sprintf("%d", limit))

		if cursor != "" {
			query.Set("after", cursor)
		}

		u, _ := url.JoinPath(c.url, "/v1/objects")
		u += "?" + query.Encode()

		resp, err := c.client.Get(u)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("bad request")
		}

		var page pageType

		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			return nil, err
		}

		for _, o := range page.Objects {
			metadata := maps.Clone(o.Properties)

			key := o.Properties["key"]
			delete(metadata, "key")

			title := o.Properties["title"]
			delete(metadata, "title")

			location := o.Properties["location"]
			delete(metadata, "location")

			content := o.Properties["content"]
			delete(metadata, "content")

			if key == "" {
				key = o.ID
			}

			d := index.Document{
				ID: key,

				Title:    title,
				Location: location,

				Content:  content,
				Metadata: metadata,
			}

			cursor = o.ID
			results = append(results, d)
		}

		if len(page.Objects) < limit {
			break
		}
	}

	slices.Reverse(results)

	return results, nil
}

func (c *Client) Index(ctx context.Context, documents ...index.Document) error {
	for _, d := range documents {
		if len(d.Embedding) == 0 && c.embedder != nil {
			embedding, err := c.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding.Data
		}

		if len(d.Embedding) == 0 {
			continue
		}

		err := c.createObject(d)

		if err != nil {
			err = c.updateObject(ctx, d)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, ids ...string) error {
	var result error

	for _, id := range ids {
		u, _ := url.JoinPath(c.url, "/v1/objects/"+c.class+"/"+convertID(id))
		req, _ := http.NewRequestWithContext(ctx, "DELETE", u, nil)

		resp, err := c.client.Do(req)

		if err != nil {
			result = errors.Join(result, err)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			if resp.StatusCode == http.StatusNotFound {
				continue
			}

			result = errors.Join(result, errors.New("unable to delete object: "+id))
		}
	}

	return result
}

func (c *Client) Query(ctx context.Context, query string, options *index.QueryOptions) ([]index.Result, error) {
	var vector strings.Builder

	embedding, err := c.embedder.Embed(ctx, query)

	if err != nil {
		return nil, err
	}

	for i, v := range embedding.Data {
		if i > 0 {
			vector.WriteString(", ")
		}

		vector.WriteString(fmt.Sprintf("%f", v))
	}

	data := executeQueryTemplate(queryData{
		Class: c.class,

		Query:  query,
		Vector: embedding.Data,

		Limit: options.Limit,
		Where: options.Filters,
	})

	body := map[string]any{
		"query": data,
	}

	u, _ := url.JoinPath(c.url, "/v1/graphql")
	resp, err := c.client.Post(u, "application/json", jsonReader(body))

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

	for _, d := range result.Data.Get[c.class] {
		key := d.Additional.ID

		metadata := map[string]string{}

		if d.Key != "" {
			key = d.Key
		}

		r := index.Result{
			Score: d.Additional.Certainty,

			Document: index.Document{
				ID: key,

				Title:    d.Title,
				Location: d.Location,

				Content:  d.Content,
				Metadata: metadata,
			},
		}

		results = append(results, r)
	}

	return results, nil
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

func (c *Client) createObject(d index.Document) error {
	properties := maps.Clone(d.Metadata)

	if properties == nil {
		properties = map[string]string{}
	}

	properties["key"] = d.ID

	properties["title"] = d.Title
	properties["location"] = d.Location

	properties["content"] = d.Content

	body := map[string]any{
		"id": convertID(d.ID),

		"class":  c.class,
		"vector": d.Embedding,

		"properties": properties,
	}

	u, _ := url.JoinPath(c.url, "/v1/objects")
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

func (c *Client) updateObject(ctx context.Context, d index.Document) error {
	properties := maps.Clone(d.Metadata)

	if properties == nil {
		properties = map[string]string{}
	}

	properties["key"] = d.ID

	properties["title"] = d.Title
	properties["location"] = d.Location

	properties["content"] = d.Content

	body := map[string]any{
		"id": convertID(d.ID),

		"class":  c.class,
		"vector": d.Embedding,

		"properties": properties,
	}

	u, _ := url.JoinPath(c.url, "/v1/objects/"+c.class+"/"+d.ID)
	req, err := http.NewRequestWithContext(ctx, "PUT", u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}

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
	Key string `json:"key"`

	Title    string `json:"title,omitempty"`
	Location string `json:"location,omitempty"`

	Content string `json:"content"`

	Additional additional `json:"_additional"`
}

type additional struct {
	ID        string  `json:"id"`
	Distance  float32 `json:"distance"`
	Certainty float32 `json:"certainty"`
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
