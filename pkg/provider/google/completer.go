package google

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/google/uuid"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
}

func NewCompleter(model string, options ...Option) (*Completer, error) {
	cfg := &Config{
		client: http.DefaultClient,

		url:   "https://generativelanguage.googleapis.com",
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Completer{
		Config: cfg,
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	body, err := convertGenerateRequest(messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		url, _ := url.JoinPath(c.url, "/v1beta/models/"+c.model+":generateContent")

		if c.token != "" {
			url += "?key=" + c.token
		}

		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("content-type", "application/json")

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, convertError(resp)
		}

		var response GenerateResponse

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		candidate := response.Candidates[0]

		content := candidate.Content.Parts[0].Text
		content = strings.TrimRight(content, "\n ")

		return &provider.Completion{
			ID:     uuid.New().String(),
			Reason: toCompletionResult(candidate.FinishReason),

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: content,
			},
		}, nil
	} else {
		defer close(options.Stream)

		url, _ := url.JoinPath(c.url, "/v1beta/models/"+c.model+":streamGenerateContent")
		url += "?alt=sse"

		if c.token != "" {
			url += "&key=" + c.token
		}

		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("content-type", "application/json")

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
			ID: uuid.New().String(),

			Message: provider.Message{
				Role: provider.MessageRoleAssistant,
			},

			//Usage: &provider.Usage{},
		}

		for i := 0; ; i++ {
			data, err := reader.ReadBytes('\n')

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			data = bytes.TrimSpace(data)

			println(string(data))

			if bytes.HasPrefix(data, []byte("event:")) {
				continue
			}

			data = bytes.TrimPrefix(data, []byte("data:"))
			data = bytes.TrimSpace(data)

			if len(data) == 0 {
				continue
			}

			var event GenerateResponse

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				return nil, err
			}

			candidate := event.Candidates[0]

			content := candidate.Content.Parts[0].Text

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			result.Message.Content += content

			options.Stream <- provider.Completion{
				ID: result.ID,
				//Reason: result.Reason,

				Message: provider.Message{
					Role: result.Message.Role,

					Content: content,
				},
			}
		}

		result.Message.Content = strings.TrimRight(result.Message.Content, "\n ")

		return result, nil
	}
}

func convertGenerateRequest(messages []provider.Message, options *provider.CompleteOptions) (*GenerateRequest, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	req := &GenerateRequest{}

	for _, m := range messages {
		switch m.Role {

		case provider.MessageRoleUser:
			content := Content{
				Role: ContentRoleUser,
			}

			content.Parts = []ContentPart{
				{
					Text: m.Content,
				},
			}

			req.Contents = append(req.Contents, content)

		case provider.MessageRoleAssistant:
			content := Content{
				Role: ContentRoleUser,
			}

			content.Parts = []ContentPart{
				{
					Text: m.Content,
				},
			}

			req.Contents = append(req.Contents, content)

		default:
			return nil, errors.New("unsupported message role")
		}
	}

	return req, nil
}

type ContentRole string

var (
	ContentRoleUser  ContentRole = "user"
	ContentRoleModel ContentRole = "model"
)

// https://ai.google.dev/gemini-api/docs/text-generation?lang=rest#chat
type GenerateRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Role ContentRole `json:"role"`

	Parts []ContentPart `json:"parts"`
}

type ContentPart struct {
	Text string `json:"text"`
}

type GenerateResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Index int `json:"index"`

	FinishReason FinishReason `json:"finishReason"`

	Content Content `json:"content"`
}

type FinishReason string

var (
	FinishReasonStop FinishReason = "STOP"
)

type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

func toCompletionResult(val FinishReason) provider.CompletionReason {
	switch val {
	case FinishReasonStop:
		return provider.CompletionReasonStop

	default:
		return ""
	}
}
