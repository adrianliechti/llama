package sentencetransformers

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
	client *http.Client

	url string
}

type Option func(*Provider)

func New(options ...Option) *Provider {
	p := &Provider{
		client: http.DefaultClient,
	}

	for _, option := range options {
		option(p)
	}

	return p
}

func WithClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

func WithURL(url string) Option {
	return func(p *Provider) {
		p.url = url
	}
}

func (*Provider) Models(ctx context.Context) ([]provider.Model, error) {
	return nil, errors.ErrUnsupported
}

func (p *Provider) Embed(ctx context.Context, model string, content string) ([]float32, error) {
	req := &vectorsRequest{
		Text: strings.TrimSpace(content),
	}

	body, _ := json.Marshal(req)
	url, _ := url.JoinPath(p.url, "/vectors")

	r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(r)

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
