package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
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

func (p *Provider) Transcribe(ctx context.Context, input provider.File, options *provider.TranscribeOptions) (*provider.Transcription, error) {
	if options == nil {
		options = &provider.TranscribeOptions{}
	}

	id := uuid.NewString()

	url, _ := url.JoinPath(p.url, "/inference")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	w.WriteField("response-format", "json")

	file, err := w.CreateFormFile("file", input.Name)

	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(file, input.Content); err != nil {
		return nil, err
	}

	w.Close()

	req, _ := http.NewRequest("POST", url, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := p.client.Do(req)

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
