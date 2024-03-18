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
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/google/uuid"
)

var (
	_ provider.Transcriber = (*Transcriber)(nil)
)

type Transcriber struct {
	*Config
}

func NewTranscriber(url string, options ...Option) (*Transcriber, error) {
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

	return &Transcriber{
		Config: cfg,
	}, nil
}

func (t *Transcriber) Transcribe(ctx context.Context, input provider.File, options *provider.TranscribeOptions) (*provider.Transcription, error) {
	if options == nil {
		options = new(provider.TranscribeOptions)
	}

	id := uuid.NewString()

	url, _ := url.JoinPath(t.url, "/inference")

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	w.WriteField("response-format", "json")

	file, err := w.CreateFormFile("file", input.Name)

	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(file, input.Content); err != nil {
		return nil, err
	}

	w.Close()

	req, _ := http.NewRequestWithContext(ctx, "POST", url, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := t.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to transcribe")
	}

	var inference InferenceResponse

	if err := json.NewDecoder(resp.Body).Decode(&inference); err != nil {
		return nil, err
	}

	content := strings.TrimSpace(inference.Text)

	if strings.EqualFold(content, "[BLANK_AUDIO]") {
		content = ""
	}

	result := provider.Transcription{
		ID: id,

		Content: content,
	}

	return &result, nil
}

// type InferenceRequest struct {
// 	Temperature *float32 `json:"temperature,omitempty"`
// }

type InferenceResponse struct {
	Text string `json:"text"`
}
