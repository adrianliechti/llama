package tesseract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/extracter"
	"github.com/adrianliechti/llama/pkg/text"
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

	url, _ := url.JoinPath(p.url, "/tesseract")

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	w.WriteField("options", p.optionsJSON())

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
		return nil, errors.New("unable to convert")
	}

	defer resp.Body.Close()

	var data Result

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Data.Stderr != "" {
		return nil, errors.New(data.Data.Stderr)
	}

	result := extracter.Document{
		Name: input.Name,
	}

	output := text.Normalize(data.Data.Stdout)

	var lines []string

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimLeft(line, ". ")
		lines = append(lines, line)
	}

	chunks := text.Split(strings.Join(lines, "\n"))

	for i, chunk := range chunks {
		chunk = strings.ReplaceAll(chunk, "\n\n", "\n")

		block := []extracter.Block{
			{
				ID:      fmt.Sprintf("%s#%d", result.Name, i),
				Content: chunk,
			},
		}

		result.Blocks = append(result.Blocks, block...)
	}

	return &result, nil
}

func (p *Provider) optionsJSON() string {
	options := map[string]any{
		"languages": []string{
			"eng",
		},
	}

	value, _ := json.Marshal(options)
	return string(value)
}
