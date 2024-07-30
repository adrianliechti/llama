package replicate

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/adrianliechti/llama/pkg/provider"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
}

func NewCompleter(options ...Option) (*Completer, error) {
	c := &Config{
		url: "https://api.replicate.com",

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	return &Completer{
		Config: c,
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	url, _ := url.JoinPath(c.url, "/v1/models/", c.model, "/predictions")
	body, err := c.convertPredictionRequest(messages, options)

	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)

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

	if options.Stream == nil {
		req, _ := http.NewRequestWithContext(ctx, "GET", prediction.URLs.Get, nil)
		req.Header.Set("Authorization", "Bearer "+c.token)

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, convertError(resp)
		}

		var prediction PredictionResponse

		if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
			return nil, err
		}

		content := strings.Join(prediction.Output, "")

		return &provider.Completion{
			ID:     prediction.ID,
			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: content,
			},
		}, nil
	} else {
		defer close(options.Stream)

		req, _ := http.NewRequestWithContext(ctx, "GET", prediction.URLs.Stream, nil)
		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "text/event-stream")

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, convertError(resp)
		}

		reader := bufio.NewReader(resp.Body)

		result := &provider.Completion{
			ID: prediction.ID,

			Message: provider.Message{
				Role: provider.MessageRoleAssistant,
			},
		}

		var currentID string
		var currentEvent string

		_ = currentID

		for i := 0; ; i++ {
			data, err := reader.ReadString('\n')

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			if strings.HasPrefix(data, "id:") {
				currentID = strings.TrimSpace(strings.TrimPrefix(data, "id:"))
				continue
			}

			if strings.HasPrefix(data, "event:") {
				currentEvent = strings.TrimSpace(strings.TrimPrefix(data, "event:"))
				continue
			}

			if currentEvent != "output" {
				continue
			}

			content := strings.TrimPrefix(data, "data:")

			if strings.TrimSpace(content) == "" {
				continue
			}

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			result.Message.Content += content

			options.Stream <- provider.Completion{
				ID:     result.ID,
				Reason: result.Reason,

				Message: provider.Message{
					Content: content,
				},
			}
		}

		return result, nil
	}
}

func (c *Completer) convertPredictionRequest(messages []provider.Message, options *provider.CompleteOptions) (*PredictionRequest, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	req := &PredictionRequest{
		Input: PredictionInput{
			StopSequences:  c.stops,
			PromptTemplate: c.template,
		},
	}

	for _, m := range messages {
		if m.Role == provider.MessageRoleSystem {
			req.Input.System = m.Content
		}

		if m.Role == provider.MessageRoleUser {
			req.Input.Prompt = m.Content
		}
	}

	return req, nil
}

type PredictionRequest struct {
	Input PredictionInput `json:"input"`
}

type PredictionResponse struct {
	ID string `json:"id"`

	Input  PredictionInput  `json:"input,omitempty"`
	Output PredictionOutput `json:"output,omitempty"`

	URLs struct {
		Cancel string `json:"cancel,omitempty"`
		Get    string `json:"get,omitempty"`
		Stream string `json:"stream,omitempty"`
	} `json:"urls,omitempty"`
}

type PredictionInput struct {
	Prompt         string `json:"prompt,omitempty"`
	PromptTemplate string `json:"prompt_template,omitempty"`

	System string `json:"system_prompt,omitempty"`

	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`

	StopSequences string `json:"stop_sequences,omitempty"`
}

type PredictionOutput []string
