package mimic

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"
)

var (
	_ provider.Synthesizer = (*Synthesizer)(nil)
)

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

	if voice := s.voice(options.Voice); voice != "" {
		query.Set("voice", voice)
	}

	u.RawQuery = query.Encode()

	req, _ := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader([]byte(content)))

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

func (s *Synthesizer) voice(val string) string {
	switch strings.ToLower(val) {

	case "en", "english":
		return "en_US/vctk_low#p239"

	case "de", "german":
		return "de_DE/thorsten_low"
	}

	return ""
}
