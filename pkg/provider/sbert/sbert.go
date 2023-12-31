package sbert

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
)

var (
	_ provider.Provider = &Provider{}
)

type Provider struct {
	url string

	client *http.Client
}

type Option func(*Provider)

func New(url string, options ...Option) (*Provider, error) {
	p := &Provider{
		url: url,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(p)
	}

	if p.url == "" {
		return nil, errors.New("invalid url")
	}

	return p, nil
}

func WithClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

func (p *Provider) Embed(ctx context.Context, model string, content string) ([]float32, error) {
	request := &vectorsRequest{
		Text: strings.TrimSpace(content),
	}

	body, _ := json.Marshal(request)
	url, _ := url.JoinPath(p.url, "/vectors")

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to vectorize text")
	}

	defer resp.Body.Close()

	var result vectorsResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Vector, nil
}

func (*Provider) Complete(ctx context.Context, model string, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	return nil, errors.ErrUnsupported
}

//curl localhost:9090/vectors -H 'Content-Type: application/json' -d '{"text": "foo bar"}'

type vectorsRequest struct {
	Text string `json:"text"`
}

type vectorsResponse struct {
	Text   string    `json:"text"`
	Vector []float32 `json:"vector"`
}
