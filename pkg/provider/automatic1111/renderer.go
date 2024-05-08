package automatic1111

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/google/uuid"
)

var _ provider.Renderer = (*Renderer)(nil)

type Renderer struct {
	*Config
}

func NewRenderer(options ...Option) (*Renderer, error) {
	c := &Config{
		url:    "http://127.0.0.1:7860",
		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	return &Renderer{
		Config: c,
	}, nil
}

func (r *Renderer) Render(ctx context.Context, input string, options *provider.RenderOptions) (*provider.Image, error) {
	body := Text2ImageRequest{
		Prompt: strings.TrimSpace(input),
		//Steps:  20,
	}

	u, _ := url.JoinPath(r.url, "/sdapi/v1/txt2img")
	resp, err := r.client.Post(u, "application/json", jsonReader(body))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, convertError(resp)
	}

	var result Text2ImageResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Images) == 0 {
		return nil, errors.New("invalid image data")
	}

	data, err := base64.StdEncoding.DecodeString(result.Images[0])

	if err != nil {
		return nil, errors.New("invalid image data")
	}

	name := strings.ReplaceAll(uuid.NewString(), "-", "") + ".png"

	image := &provider.Image{
		Name:    name,
		Content: io.NopCloser(bytes.NewReader(data)),
	}

	return image, nil
}

type Text2ImageRequest struct {
	Prompt string `json:"prompt"`
	Steps  int    `json:"steps,omitempty"`
}

type Text2ImageResponse struct {
	Images []string `json:"images"`
}
