package langchain

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

var (
	_ provider.Completer = (*Client)(nil)
)

type Client struct {
	url string

	client *http.Client
}

type Option func(*Client)

func New(url string, options ...Option) (*Client, error) {
	if url == "" {
		return nil, errors.New("invalid url")
	}

	c := &Client{
		url: url,

		client: http.DefaultClient,
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

func (c *Client) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	id := uuid.NewString()

	body, err := c.convertRunInput(messages, options)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		url, _ := url.JoinPath(c.url, "invoke")
		resp, err := c.client.Post(url, "application/json", jsonReader(body))

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("unable to complete")
		}

		defer resp.Body.Close()

		result := provider.Completion{
			ID: id,

			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: "hoi",
			},
		}

		return &result, nil
	} else {
		defer close(options.Stream)

		url, _ := url.JoinPath(c.url, "stream")
		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("unable to complete")
		}

		reader := bufio.NewReader(resp.Body)

		var resultText strings.Builder
		resultRole := provider.MessageRoleAssistant
		resultReason := provider.CompletionReason("")

		for i := 0; ; i++ {
			data, err := reader.ReadBytes('\n')

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				return nil, err
			}

			data = bytes.TrimSpace(data)

			println(string(data))

			if bytes.EqualFold(data, []byte("event: end")) {
				resultReason = provider.CompletionReasonStop

				options.Stream <- provider.Completion{
					ID: id,

					Reason: resultReason,

					Message: provider.Message{
						Role:    resultRole,
						Content: "",
					},
				}
			}

			if bytes.HasPrefix(data, []byte("event: ")) {
				continue
			}

			data = bytes.TrimPrefix(data, []byte("data: "))

			if len(data) == 0 {
				continue
			}

			var run RunData

			if err := json.Unmarshal([]byte(data), &run); err != nil {
				return nil, err
			}

			var content = run.Output

			if content == "" {
				content = run.Content
			}

			if i == 0 {
				content = strings.TrimLeftFunc(content, unicode.IsSpace)
			}

			if content == "" {
				continue
			}

			resultText.WriteString(content)
			resultRole = toMessageRole(run)

			options.Stream <- provider.Completion{
				ID: id,

				Reason: resultReason,

				Message: provider.Message{
					Role:    resultRole,
					Content: content,
				},
			}
		}

		result := provider.Completion{
			ID: id,

			Reason: resultReason,

			Message: provider.Message{
				Role:    resultRole,
				Content: resultText.String(),
			},
		}

		return &result, nil
	}
}

func (c *Client) convertRunInput(messages []provider.Message, options *provider.CompleteOptions) (*RunInput, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	if len(messages) == 0 {
		return nil, errors.New("no messages")
	}

	if messages[len(messages)-1].Role != provider.MessageRoleUser {
		return nil, errors.New("last message must be from user")
	}

	var input string
	var history []Message

	for _, m := range messages {
		messsage := Message{
			Type:    toRole(m.Role),
			Content: strings.TrimSpace(m.Content),
		}

		if messsage.Type == "" {
			continue
		}

		history = append(history, messsage)
	}

	input = history[len(history)-1].Content
	history = history[:len(history)-1]

	result := &RunInput{
		Input: Input{
			Input:   input,
			History: history,
		},
	}

	return result, nil
}

func toRole(role provider.MessageRole) MessageType {
	switch role {
	case provider.MessageRoleSystem:
		return MessageTypeHuman
	case provider.MessageRoleUser:
		return MessageTypeHuman
	case provider.MessageRoleAssistant:
		return MessageTypeAI
	default:
		return ""
	}
}

func toMessageRole(run RunData) provider.MessageRole {
	return provider.MessageRoleAssistant
}

func jsonReader(v any) io.Reader {
	b := new(bytes.Buffer)

	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)

	enc.Encode(v)
	return b
}