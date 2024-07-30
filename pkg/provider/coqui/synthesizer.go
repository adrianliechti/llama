package coqui

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
)

var _ provider.Synthesizer = (*Synthesizer)(nil)

type Synthesizer struct {
	*Config
}

func NewSynthesizer(url string, options ...Option) (*Synthesizer, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	cfg := &Config{
		url: url,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Synthesizer{
		Config: cfg,
	}, nil
}

func (s *Synthesizer) Synthesize(ctx context.Context, content string, options *provider.SynthesizeOptions) (*provider.Synthesis, error) {
	if options == nil {
		options = new(provider.SynthesizeOptions)
	}

	u, _ := url.Parse(strings.TrimRight(s.url, "/") + "/api/tts")

	query := u.Query()

	query.Set("text", content)
	query.Set("speaker_id", "p376")

	u.RawQuery = query.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)

	resp, err := s.client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	id := uuid.NewString()

	return &provider.Synthesis{
		ID: id,

		Name:    id + ".wav",
		Content: resp.Body,
	}, nil
}
