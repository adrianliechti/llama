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
	"github.com/adrianliechti/llama/pkg/to"

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

		return &provider.Completion{
			ID:     uuid.New().String(),
			Reason: toCompletionResult(candidate.FinishReason),

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: toContent(candidate.Content),

				ToolCalls: toToolCalls(candidate.Content),
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

		resultToolCalls := map[string]provider.ToolCall{}

		for i := 0; ; i++ {
			data, err := reader.ReadBytes('\n')

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			data = bytes.TrimSpace(data)

			if !bytes.HasPrefix(data, []byte("data:")) {
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

			content := toContent(candidate.Content)

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			result.Message.Content += content

			if len(content) > 0 {
				options.Stream <- provider.Completion{
					ID: result.ID,

					Message: provider.Message{
						Role:    provider.MessageRoleAssistant,
						Content: content,
					},
				}
			}

			for _, c := range toToolCalls(candidate.Content) {
				resultToolCalls[c.Name] = c
			}
		}

		result.Message.Content = strings.TrimRightFunc(result.Message.Content, unicode.IsSpace)

		if len(resultToolCalls) > 0 {
			result.Message.ToolCalls = to.Values(resultToolCalls)
		}

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
			parts := []ContentPart{
				{
					Text: m.Content,
				},
			}

			req.Contents = append(req.Contents, Content{
				Role:  ContentRoleUser,
				Parts: parts,
			})

		case provider.MessageRoleAssistant:
			var parts []ContentPart

			if m.Content != "" {
				parts = append(parts, ContentPart{
					Text: m.Content,
				})
			}

			for _, c := range m.ToolCalls {
				parts = append(parts, ContentPart{
					FunctionCall: &FunctionCall{
						Name: c.Name,
						Args: json.RawMessage([]byte(c.Arguments)),
					},
				})
			}

			req.Contents = append(req.Contents, Content{
				Role:  ContentRoleModel,
				Parts: parts,
			})

		case provider.MessageRoleTool:
			parts := []ContentPart{
				{
					FunctionResponse: &FunctionResponse{
						Name: m.Tool,

						Response: Response{
							Name:    m.Tool,
							Content: json.RawMessage([]byte(m.Content)),
						},
					},
				},
			}

			req.Contents = append(req.Contents, Content{
				Role:  ContentRoleUser,
				Parts: parts,
			})

		default:
			return nil, errors.New("unsupported message role")
		}
	}

	var functions []FunctionDeclaration

	for _, t := range options.Tools {
		function := FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,

			Parameters: t.Parameters,
		}

		functions = append(functions, function)
	}

	if len(functions) > 0 {
		req.Tools = []Tool{
			{
				FunctionDeclarations: functions,
			},
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

	Tools []Tool `json:"tools,omitempty"`
}

type Content struct {
	Role ContentRole `json:"role"`

	Parts []ContentPart `json:"parts"`
}

type ContentPart struct {
	Text string `json:"text,omitempty"`

	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
}

type FunctionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type FunctionResponse struct {
	Name string `json:"name"`

	Response Response `json:"response,omitempty"`
}

type Response struct {
	Name string `json:"name"`

	Content json.RawMessage `json:"content"`
}

type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"function_declarations,omitempty"`
}

type FunctionDeclaration struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	Parameters map[string]any `json:"parameters"`
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

func toContent(content Content) string {
	for _, p := range content.Parts {
		if p.Text == "" {
			continue
		}

		return p.Text
	}

	return ""
}

func toToolCalls(content Content) []provider.ToolCall {
	var result []provider.ToolCall

	for _, p := range content.Parts {
		if p.FunctionCall == nil {
			continue
		}

		arguments, _ := p.FunctionCall.Args.MarshalJSON()

		call := provider.ToolCall{
			ID: uuid.NewString(),

			Name:      p.FunctionCall.Name,
			Arguments: string(arguments),
		}

		result = append(result, call)
	}

	return result
}

func toCompletionResult(val FinishReason) provider.CompletionReason {
	switch val {
	case FinishReasonStop:
		return provider.CompletionReasonStop

	default:
		return ""
	}
}
