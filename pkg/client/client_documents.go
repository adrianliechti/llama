package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type Document struct {
	ID string `json:"id,omitempty"`

	Title   string `json:"title,omitempty"`
	Source  string `json:"source,omitempty"`
	Content string `json:"content,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`

	Embedding []float32 `json:"embedding,omitempty"`
}

type DocumentService struct {
	Options []RequestOption
}

func NewDocumentService(opts ...RequestOption) *DocumentService {
	return &DocumentService{
		Options: opts,
	}
}

func (r *DocumentService) New(ctx context.Context, index string, documents []Document, opts ...RequestOption) ([]Document, error) {
	c := newRequestConfig(append(r.Options, opts...)...)

	var data bytes.Buffer

	if err := json.NewEncoder(&data).Encode(documents); err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/v1/index/"+index, &data)
	req.Header.Set("Content-Type", "application/json")

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return nil, errors.New(resp.Status)
	}

	return documents, nil
}

func (r *DocumentService) List(ctx context.Context, index string, opts ...RequestOption) ([]Document, error) {
	c := newRequestConfig(append(r.Options, opts...)...)

	req, _ := http.NewRequestWithContext(ctx, "GET", c.URL+"/v1/index/"+index, nil)

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var documents []Document

	if err := json.NewDecoder(resp.Body).Decode(&documents); err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *DocumentService) Delete(ctx context.Context, index string, ids []string, opts ...RequestOption) error {
	c := newRequestConfig(append(r.Options, opts...)...)

	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(ids); err != nil {
		return err
	}

	req, _ := http.NewRequestWithContext(ctx, "DELETE", c.URL+"/v1/index/"+index, &body)
	req.Header.Set("Content-Type", "application/json")

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.New(resp.Status)
	}

	return nil
}
