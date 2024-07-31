package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
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
		url: "https://api.anthropic.com",

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

	if len(options.Tools) > 0 {
		return nil, errors.New("tools are not yet supported")
	}

	url, _ := url.JoinPath(c.url, "/v1/messages")
	body, err := convertChatRequest(c.model, messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("x-api-key", c.token)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("content-type", "application/json")

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, convertError(resp)
		}

		var response MessageResponse

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		if response.Role != MessageRoleAssistant || len(response.Content) != 1 {
			return nil, errors.New("invalid complete response")
		}

		role := provider.MessageRoleAssistant
		content := response.Content[0].Text

		return &provider.Completion{
			ID:     response.ID,
			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    role,
				Content: content,
			},

			Usage: &provider.Usage{
				InputTokens:  response.Usage.InputTokens,
				OutputTokens: response.Usage.OutputTokens,
			},
		}, nil
	} else {
		defer close(options.Stream)

		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("x-api-key", c.token)
		req.Header.Set("anthropic-version", "2023-06-01")
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
			Message: provider.Message{
				Role: provider.MessageRoleAssistant,
			},

			Usage: &provider.Usage{},
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

			if bytes.HasPrefix(data, []byte("event:")) {
				continue
			}

			data = bytes.TrimPrefix(data, []byte("data:"))
			data = bytes.TrimSpace(data)

			if len(data) == 0 {
				continue
			}

			var event MessageEvent

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				return nil, err
			}

			if event.Message != nil {
				result.ID = event.Message.ID

				if event.Message.Usage.InputTokens > 0 {
					result.Usage.InputTokens = event.Message.Usage.InputTokens
				}

				if event.Message.Usage.OutputTokens > 0 {
					result.Usage.OutputTokens = event.Message.Usage.OutputTokens
				}
			}

			if event.Type == EventTypeMessageStop {
				result.Reason = provider.CompletionReasonStop
				break
			}

			if event.Usage.InputTokens > 0 {
				result.Usage.InputTokens = event.Usage.InputTokens
			}

			if event.Usage.OutputTokens > 0 {
				result.Usage.OutputTokens = event.Usage.OutputTokens
			}

			if event.Type != EventTypeContentBlockDelta || event.Delta == nil {
				continue
			}

			var content = event.Delta.Text

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			result.Message.Content += content

			options.Stream <- provider.Completion{
				ID:     result.ID,
				Reason: result.Reason,

				Message: provider.Message{
					Role:    result.Message.Role,
					Content: content,
				},
			}
		}

		if result.Usage.OutputTokens == 0 {
			result.Usage = nil
		}

		return result, nil
	}
}

func convertChatRequest(model string, messages []provider.Message, options *provider.CompleteOptions) (*MessagesRequest, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	stream := options.Stream != nil

	req := &MessagesRequest{
		Model:  model,
		Stream: stream,

		MaxTokens: 1024,
	}

	if options.Stop != nil {
		req.StopSequences = options.Stop
	}

	if options.MaxTokens != nil {
		req.MaxTokens = *options.MaxTokens
	}

	if options.Temperature != nil {
		req.Temperature = options.Temperature
	}

	for _, m := range messages {
		switch m.Role {
		case provider.MessageRoleSystem:
			req.System = m.Content

		case provider.MessageRoleUser:
			message := Message{
				Role:    MessageRoleUser,
				Content: m.Content,
			}

			if len(m.Files) > 0 {
				message.Content = ""

				message.Contents = []Content{
					{
						Type: ContentTypeText,
						Text: m.Content,
					},
				}

				for _, f := range m.Files {
					data, err := io.ReadAll(f.Content)

					if err != nil {
						return nil, err
					}

					message.Contents = append(message.Contents, Content{
						Type: ContentTypeImage,

						Source: &ContentSource{
							Type: "base64",

							MediaType: http.DetectContentType(data),
							Data:      base64.StdEncoding.EncodeToString(data),
						},
					})
				}
			}

			req.Messages = append(req.Messages, message)

		case provider.MessageRoleAssistant:
			req.Messages = append(req.Messages, Message{
				Role:    MessageRoleAssistant,
				Content: m.Content,
			})

		default:
			return nil, errors.New("unsupported message role")
		}
	}

	return req, nil
}

type MessageRole string

var (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type MessagesRequest struct {
	Model string `json:"model"`

	Stream bool   `json:"stream"`
	System string `json:"system,omitempty"`

	Messages []Message `json:"messages"`

	MaxTokens     int      `json:"max_tokens,omitempty"`
	Temperature   *float32 `json:"temperature,omitempty"`
	StopSequences []string `json:"stop_sequences,omitempty"`
}

type Message struct {
	Role MessageRole `json:"role"`

	Content  string    `json:"content"`
	Contents []Content `json:"contents,omitempty"`
}

func (m *Message) MarshalJSON() ([]byte, error) {
	if m.Content != "" && m.Contents != nil {
		return nil, errors.New("cannot have both content and contents")
	}

	if len(m.Contents) > 0 {
		msg := struct {
			Role MessageRole `json:"role"`

			Content  string    `json:"-"`
			Contents []Content `json:"content,omitempty"`
		}(*m)

		return json.Marshal(msg)
	}

	msg := struct {
		Role MessageRole `json:"role"`

		Content  string    `json:"content"`
		Contents []Content `json:"-"`
	}(*m)

	return json.Marshal(msg)
}

func (m *Message) UnmarshalJSON(data []byte) error {
	m1 := struct {
		Role MessageRole `json:"role"`

		Content  string `json:"content"`
		Contents []Content
	}{}

	if err := json.Unmarshal(data, &m1); err == nil {
		*m = Message(m1)
		return nil
	}

	m2 := struct {
		Role MessageRole `json:"role"`

		Content  string
		Contents []Content `json:"content"`
	}{}

	if err := json.Unmarshal(data, &m2); err == nil {
		*m = Message(m2)
		return err
	}

	return nil
}

type ContentType string

var (
	ContentTypeText      ContentType = "text"
	ContentTypeTextDelta ContentType = "text_delta"
	ContentTypeImage     ContentType = "image"
)

type Content struct {
	Type ContentType `json:"type"`

	Text   string         `json:"text,omitempty"`
	Source *ContentSource `json:"source,omitempty"`
}

type ContentSource struct {
	Type string `json:"type"`

	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type ResponseType string

var (
	ResponseTypeMessage ResponseType = "message"
)

type StopReason string

var (
	StopReasonEndTurn      StopReason = "end_turn"
	StopReasonMaxTokens    StopReason = "max_tokens"
	StopReasonStopSequence StopReason = "stop_sequence"
)

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type MessageResponse struct {
	ID string `json:"id"`

	Type  ResponseType `json:"type"`
	Model string       `json:"model"`

	Role MessageRole `json:"role"`

	Content []Content `json:"content"`

	StopReason   StopReason `json:"stop_reason,omitempty"`
	StopSequence []string   `json:"stop_sequence,omitempty"`

	Usage Usage `json:"usage"`
}

type MessageDelta struct {
	Text string `json:"text,omitempty"`

	StopReason   StopReason `json:"stop_reason,omitempty"`
	StopSequence []string   `json:"stop_sequence,omitempty"`
}

type EventType string

var (
	EventTypePing EventType = "ping"

	EventTypeMessageStart EventType = "message_start"
	EventTypeMessageDelta EventType = "message_delta"
	EventTypeMessageStop  EventType = "message_stop"

	EventTypeContentBlockStart EventType = "content_block_start"
	EventTypeContentBlockDelta EventType = "content_block_delta"
	EventTypeContentBlockStop  EventType = "content_block_stop"
)

type MessageEvent struct {
	Type EventType `json:"type"`

	Index int `json:"index"`

	Message *MessageResponse `json:"message,omitempty"`
	Delta   *MessageDelta    `json:"delta,omitempty"`

	Usage Usage `json:"usage"`
}
