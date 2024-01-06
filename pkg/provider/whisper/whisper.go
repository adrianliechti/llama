package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/google/uuid"
)

var (
	_ provider.Transcriber = (*Provider)(nil)
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

func (p *Provider) Transcribe(ctx context.Context, input any, options *provider.TranscribeOptions) (*provider.Transcription, error) {
	if options == nil {
		options = &provider.TranscribeOptions{}
	}

	id := uuid.NewString()

	url, _ := url.JoinPath(p.url, "/inference")
	body, err := p.convertInferenceRequest(input, options)

	if err != nil {
		return nil, err
	}

	resp, err := p.client.Post(url, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to transcribe")
	}

	defer resp.Body.Close()

	var inference InferenceResponse

	if err := json.NewDecoder(resp.Body).Decode(&inference); err != nil {
		return nil, err
	}

	result := provider.Transcription{
		ID: id,

		Content: inference.Text,
	}

	return &result, nil
}

func (p *Provider) convertInferenceRequest(input any, options *provider.TranscribeOptions) (*InferenceRequest, error) {
	if options == nil {
		options = &provider.TranscribeOptions{}
	}

	req := &InferenceRequest{
		Temperature: options.Temperature,
	}

	return req, nil
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}
