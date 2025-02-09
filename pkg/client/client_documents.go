package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type DocumentRequest struct {
	Index     string
	Documents []Document
}

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

func (r *DocumentService) New(ctx context.Context, body DocumentRequest, opts ...RequestOption) ([]Document, error) {
	c := newRequestConfig(append(r.Options, opts...)...)

	var data bytes.Buffer

	if err := json.NewEncoder(&data).Encode(body.Documents); err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", c.URL+"/v1/index", &data)
	req.Header.Set("Content-Type", "application/json")

	if body.Index != "" {
		req.URL.Path = req.URL.Path + "/" + body.Index
	}

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

	return body.Documents, nil
}
