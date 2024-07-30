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

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
}

func NewCompleter(url string, options ...Option) (*Completer, error) {
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

	return &Completer{
		Config: cfg,
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
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

		req, _ := http.NewRequestWithContext(ctx, "POST", url, jsonReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, convertError(resp)
		}

		defer resp.Body.Close()

		return &provider.Completion{
			ID:     id,
			Reason: provider.CompletionReasonStop,

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: "hoi",
			},
		}, nil
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
			return nil, convertError(resp)
		}

		reader := bufio.NewReader(resp.Body)

		var resultText strings.Builder
		resultRole := provider.MessageRoleAssistant
		resultReason := provider.CompletionReason("")

		for i := 0; ; i++ {
			data, err := reader.ReadBytes('\n')

			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			data = bytes.TrimSpace(data)

			if bytes.EqualFold(data, []byte("event: end")) {
				resultReason = provider.CompletionReasonStop

				options.Stream <- provider.Completion{
					ID:     id,
					Reason: resultReason,

					Message: provider.Message{
						Role:    resultRole,
						Content: "",
					},
				}
			}

			if bytes.HasPrefix(data, []byte("event:")) {
				continue
			}

			data = bytes.TrimPrefix(data, []byte("data:"))
			data = bytes.TrimSpace(data)

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
				ID:     id,
				Reason: resultReason,

				Message: provider.Message{
					Role:    resultRole,
					Content: content,
				},
			}
		}

		return &provider.Completion{
			ID:     id,
			Reason: resultReason,

			Message: provider.Message{
				Role:    resultRole,
				Content: resultText.String(),
			},
		}, nil
	}
}

func (c *Completer) convertRunInput(messages []provider.Message, options *provider.CompleteOptions) (*RunInput, error) {
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
