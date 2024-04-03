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
	url string

	client   *http.Client
	embedder index.Embedder

	class string
}

type Option func(*Client)

func New(url, namespace string, options ...Option) (*Client, error) {
	c := &Client{
		url: url,

		client: http.DefaultClient,

		class: namespace,
	}

	for _, option := range options {
		option(c)
	}

	if c.embedder == nil {
		return nil, errors.New("embedder is required")
	}

	return c, nil
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
			id := o.ID
			content := o.Properties["content"]

			metadata := maps.Clone(o.Properties)
			delete(metadata, "content")

			d := index.Document{
				ID:      id,
				Content: content,
			}

			cursor = id
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
		d.ID = generateID(d)

		if len(d.Embedding) == 0 && c.embedder != nil {
			embedding, err := c.embedder.Embed(ctx, d.Content)

			if err != nil {
				return err
			}

			d.Embedding = embedding
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
		id := convertID(id)

		u, _ := url.JoinPath(c.url, "/v1/objects/"+c.class+"/"+id)
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

	for i, v := range embedding {
		if i > 0 {
			vector.WriteString(", ")
		}

		vector.WriteString(fmt.Sprintf("%f", v))
	}

	data := executeQueryTemplate(queryData{
		Class: c.class,

		Query:  query,
		Vector: embedding,

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
		title := d.Additional.ID
		location := d.Additional.ID

		metadata := map[string]string{}

		if d.FileName != "" {
			metadata["filename"] = d.FileName
			title = d.FileName
			location = d.FileName
		}

		if d.FilePart != "" {
			metadata["filepart"] = d.FilePart

			if location != "" {
				location += "#" + d.FilePart
			}
		}

		r := index.Result{
			Document: index.Document{
				ID: d.Additional.ID,

				Title:    title,
				Content:  d.Content,
				Location: location,

				Metadata: metadata,
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

	return convertID(d.ID)
}

func convertID(id string) string {
	return uuid.NewMD5(uuid.NameSpaceOID, []byte(id)).String()
}

func (c *Client) createObject(d index.Document) error {
	properties := maps.Clone(d.Metadata)

	if properties == nil {
		properties = map[string]string{}
	}

	properties["content"] = d.Content

	filename := d.Metadata["filename"]
	filepart := d.Metadata["filepart"]

	if filename == "" {
		filename = d.ID

		if d.Location != "" {
			filename = d.Location
		}
	}

	if filepart == "" {
		filepart = "0"
	}

	properties["filename"] = filename
	properties["filepart"] = filepart

	body := map[string]any{
		"id": d.ID,

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

	properties["content"] = d.Content

	body := map[string]any{
		"id": d.ID,

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
	Content string `json:"content"`

	// HACK
	FileName string `json:"filename,omitempty"`
	FilePart string `json:"filepart,omitempty"`

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
