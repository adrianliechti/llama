package unstructured

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/adrianliechti/llama/pkg/extracter"
)

var _ extracter.Provider = &Provider{}

type Provider struct {
	url string

	client *http.Client
}

type Option func(*Provider)

func New(options ...Option) (*Provider, error) {
	p := &Provider{
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

func WithURL(url string) Option {
	return func(p *Provider) {
		p.url = url
	}
}

func (p *Provider) Extract(ctx context.Context, input extracter.File, options *extracter.ExtractOptions) (*extracter.Document, error) {
	if options == nil {
		options = &extracter.ExtractOptions{}
	}

	url, _ := url.JoinPath(p.url, "/general/v0/general")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	file, err := w.CreateFormFile("files", input.Name)

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
		return nil, errors.New("unable to convert")
	}

	defer resp.Body.Close()

	var elements []Element

	if err := json.NewDecoder(resp.Body).Decode(&elements); err != nil {
		return nil, err
	}

	result := extracter.Document{}

	page := extracter.Page{}

	for _, e := range elements {
		if e.Metadata.PageNumber > len(result.Pages) {
			if len(page.Blocks) > 0 {
				result.Pages = append(result.Pages, page)
				page = extracter.Page{}
			}
		}

		block := extracter.Block{
			Text: e.Text,
		}

		page.Blocks = append(page.Blocks, block)
	}

	return &result, nil
}
