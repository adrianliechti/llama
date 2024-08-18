package flux

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/adrianliechti/llama/pkg/provider"
)

type Renderer struct {
	*Config
}

func NewRenderer(options ...Option) (*Renderer, error) {
	c := NewConfig(options...)

	if c.model == "" {
		return nil, errors.New("model is required")
	}

	return &Renderer{
		Config: c,
	}, nil
}

func (r *Renderer) Render(ctx context.Context, input string, options *provider.RenderOptions) (*provider.Image, error) {
	if options == nil {
		options = new(provider.RenderOptions)
	}

	body, err := convertPredictionRequest(r.Config, input, options)

	if err != nil {
		return nil, err
	}

	url, _ := url.JoinPath(r.url, "/v1/models/"+r.model+"/predictions")

	req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
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

		output, err := convertPredictionResponse(r.Config, prediction)

		if err != nil {
			return nil, err
		}

		return output, nil
	}
}

func convertPredictionRequest(cfg *Config, input string, options *provider.RenderOptions) (*CreatePredictionRequest, error) {
	if options == nil {
		options = new(provider.RenderOptions)
	}

	return &CreatePredictionRequest{
		Input: map[string]any{
			"prompt": input,

			"aspect_ratio": "1:1",

			"safety_tolerance":       5,
			"disable_safety_checker": true,
		},
	}, nil
}

func convertPredictionResponse(cfg *Config, prediction PredictionResponse) (*provider.Image, error) {
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
	req.Header.Set("Authorization", "Bearer "+cfg.token)

	resp, err := cfg.client.Do(req)

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
