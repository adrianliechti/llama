package flux

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
	"slices"
	"time"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Renderer struct {
	*Config
}

func NewRenderer(url, model string, options ...Option) (*Renderer, error) {
	if url == "" {
		url = "https://api.replicate.com/"
	}

	if !slices.Contains(SupportedModels, model) {
		return nil, errors.New("unsupported model")
	}

	cfg := &Config{
		client: http.DefaultClient,

		url:   url,
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Renderer{
		Config: cfg,
	}, nil
}

func (r *Renderer) Render(ctx context.Context, input string, options *provider.RenderOptions) (*provider.Image, error) {
	if options == nil {
		options = new(provider.RenderOptions)
	}

	body, err := r.convertPredictionRequest(input, options)

	if err != nil {
		return nil, err
	}

	u, _ := url.JoinPath(r.url, "/v1/models/"+r.model+"/predictions")

	if body.Version != "" {
		u, _ = url.JoinPath(r.url, "/v1/predictions")
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", u, jsonReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.token)

	resp, err := r.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, convertError(resp)
	}

	var prediction PredictionResponse

	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return nil, err
	}

	for {
		req, _ := http.NewRequestWithContext(ctx, "GET", prediction.URL.Get, nil)
		req.Header.Set("Authorization", "Bearer "+r.token)

		resp, err := r.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, convertError(resp)
		}

		if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
			return nil, err
		}

		if prediction.Status == PredictionStatusStarting || prediction.Status == PredictionStatusProcessing {
			time.Sleep(5 * time.Second)
			continue
		}

		if prediction.Status != PredictionStatusSucceeded {
			return nil, errors.New("prediction " + string(prediction.Status))
		}

		output, err := r.convertPredictionResponse(prediction)

		if err != nil {
			return nil, err
		}

		return output, nil
	}
}

func (r *Renderer) convertPredictionRequest(prompt string, options *provider.RenderOptions) (*CreatePredictionRequest, error) {
	if options == nil {
		options = new(provider.RenderOptions)
	}

	var input map[string]any

	switch r.model {
	case FluxSchnell:
		// https://replicate.com/black-forest-labs/flux-schnell/api/schema#input-schema
		input = map[string]any{
			"prompt": prompt,

			"aspect_ratio":  "3:2",
			"output_format": "png",

			"disable_safety_checker": true,
		}

	case FluxDev:
		// https://replicate.com/black-forest-labs/flux-dev/api/schema#input-schema
		input = map[string]any{
			"prompt": prompt,

			"aspect_ratio":  "3:2",
			"output_format": "png",

			"disable_safety_checker": true,
		}

	case FluxPro:
		// https://replicate.com/black-forest-labs/flux-pro/api/schema#input-schema
		input = map[string]any{
			"prompt": prompt,

			"aspect_ratio":  "3:2",
			"output_format": "png",

			"safety_tolerance": 6,
		}

	case FluxPro11:
		// https://replicate.com/black-forest-labs/flux-1.1-pro/api/schema#input-schema
		input = map[string]any{
			"prompt": prompt,

			"aspect_ratio":  "3:2",
			"output_format": "png",

			"safety_tolerance": 6,
		}

	case FluxProUltra11:
		// https://replicate.com/black-forest-labs/flux-1.1-pro-ultra/api/schema#input-schema
		input = map[string]any{
			"prompt": prompt,

			"aspect_ratio":  "3:2",
			"output_format": "png",

			"safety_tolerance": 6,
		}

		if options.Style == provider.ImageStyleNatural {
			input["raw"] = true
		}

	case FluxDevRealism:
		// https://replicate.com/xlabs-ai/flux-dev-realism/api/schema#input-schema
		input = map[string]any{
			"prompt": prompt,

			"aspect_ratio":  "3:2",
			"output_format": "png",
		}
	}

	if len(input) == 0 {
		return nil, errors.New("unsupported model")
	}

	return &CreatePredictionRequest{
		Version: ModelVersion[r.model],

		Input: input,
	}, nil
}

func (r *Renderer) convertPredictionResponse(prediction PredictionResponse) (*provider.Image, error) {
	var url string
	var urls []string

	json.Unmarshal(prediction.Output, &url)
	json.Unmarshal(prediction.Output, &urls)

	if len(urls) > 0 {
		url = urls[0]
	}

	if url == "" {
		return nil, errors.New("invalid output")
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+r.token)

	resp, err := r.client.Do(req)

	if err != nil {
		return nil, err
	}

	return &provider.Image{
		ID: prediction.ID,

		Name:    path.Base(url),
		Content: resp.Body,
	}, nil
}

type CreatePredictionRequest struct {
	Version string `json:"version,omitempty"`

	Input any `json:"input"`
}

type PredictionStatus string

const (
	PredictionStatusStarting   PredictionStatus = "starting"
	PredictionStatusProcessing PredictionStatus = "processing"
	PredictionStatusSucceeded  PredictionStatus = "succeeded"
	PredictionStatusFailed     PredictionStatus = "failed"
	PredictionStatusCanceled   PredictionStatus = "canceled"
)

type PredictionResponse struct {
	ID string `json:"id"`

	Model   string `json:"model"`
	Version string `json:"version"`

	Input  json.RawMessage `json:"input"`
	Output json.RawMessage `json:"output"`

	Status PredictionStatus `json:"status"`

	CreatedAt   time.Time `json:"created_at"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`

	URL struct {
		Get    string `json:"get"`
		Cancel string `json:"cancel"`
	} `json:"urls"`
}
