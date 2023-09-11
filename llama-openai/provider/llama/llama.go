package llama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type Provider struct {
	client *http.Client

	url string

	username string
	password string
}

var (
	headerData  = []byte("data: ")
	errorPrefix = []byte(`data: {"error":`)
)

func FromEnvironment() (*Provider, error) {
	u, err := url.Parse(os.Getenv("LLAMA_URL"))

	if err != nil {
		return nil, err
	}

	if u.Host == "" {
		return nil, errors.New("LLAMA_URL is not set")
	}

	var username string
	var password string

	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()

		u.User = nil
	}

	client := http.DefaultClient

	return &Provider{
		client: client,

		url: u.String(),

		username: username,
		password: password,
	}, nil
}

func (p *Provider) Models(ctx context.Context) ([]openai.Model, error) {
	return []openai.Model{
		{
			ID: "default",

			Object: "model",
			Root:   "default",

			OwnedBy:   "owner",
			CreatedAt: time.Now().Unix(),
		},
	}, nil
}

func (p *Provider) Chat(ctx context.Context, request openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	sessionID := uuid.New().String()

	req, err := convertRequest(request)

	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(req)
	url, _ := url.JoinPath(p.url, "/completion")

	r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Cache-Control", "no-cache")

	if p.password != "" {
		r.SetBasicAuth(p.username, p.password)
	}

	resp, err := p.client.Do(r)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unable to complete chat")
	}

	defer resp.Body.Close()

	var result completionResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	model := result.Model

	if model == "" {
		model = request.Model
	}

	return &openai.ChatCompletionResponse{
		ID: sessionID,

		Model:  model,
		Object: "chat.completion",

		Created: time.Now().Unix(),

		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Content: strings.TrimSpace(result.Content),
				},

				FinishReason: openai.FinishReasonStop,
			},
		},
	}, nil
}

func (p *Provider) ChatStream(ctx context.Context, request openai.ChatCompletionRequest, stream chan<- openai.ChatCompletionStreamResponse) error {
	sessionID := uuid.New().String()

	req, err := convertRequest(request)

	if err != nil {
		return err
	}

	data, _ := json.Marshal(req)
	url, _ := url.JoinPath(p.url, "/completion")

	r, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "text/event-stream")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Cache-Control", "no-cache")

	if err != nil {
		return err
	}

	resp, err := p.client.Do(r)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("unable to complete chat")
	}

	reader := bufio.NewReader(resp.Body)

	for {
		data, err := reader.ReadBytes('\n')

		if err != nil {
			return err
		}

		data = bytes.TrimSpace(data)

		// if bytes.HasPrefix(data, errorPrefix) {
		// }

		data = bytes.TrimPrefix(data, headerData)

		if len(data) == 0 {
			continue
		}

		var result completionResponse

		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return err
		}

		model := result.Model

		if model == "" {
			model = request.Model
		}

		status := openai.FinishReasonNull

		if result.Stop {
			status = openai.FinishReasonStop
		}

		stream <- openai.ChatCompletionStreamResponse{
			ID: sessionID,

			Model:  model,
			Object: "chat.completion.chunk",

			Created: time.Now().Unix(),

			Choices: []openai.ChatCompletionStreamChoice{
				{
					Delta: openai.ChatCompletionStreamChoiceDelta{
						Role:    openai.ChatMessageRoleAssistant,
						Content: result.Content,
					},

					FinishReason: status,
				},
			},
		}

		if result.Stop {
			break
		}
	}

	return nil
}

func verifyRequest(req openai.ChatCompletionRequest) error {
	messages := slices.Clone(req.Messages)

	if len(messages) == 0 {
		return errors.New("there must be at least one message")
	}

	if messages[0].Role == openai.ChatMessageRoleSystem {
		messages = messages[1:]
	}

	errRole := errors.New("model only supports 'system', 'user' and 'assistant' roles, starting with 'system', then 'user' and alternating (u/a/u/a/u...)")

	for i, m := range messages {
		if m.Role != openai.ChatMessageRoleUser && m.Role != openai.ChatMessageRoleAssistant {
			return errRole
		}

		if (i+1)%2 == 1 && m.Role != openai.ChatMessageRoleUser {
			return errRole
		}

		if (i+1)%2 == 0 && m.Role != openai.ChatMessageRoleAssistant {
			return errRole
		}
	}

	if messages[len(messages)-1].Role != openai.ChatMessageRoleUser {
		return errors.New("last message must be from user")
	}

	return nil
}

func convertRequest(req openai.ChatCompletionRequest) (*completionRequest, error) {
	if err := verifyRequest(req); err != nil {
		return nil, err
	}

	messages := slices.Clone(req.Messages)

	system := "You are a helpful, respectful and honest assistant. Always answer as helpfully as possible, while being safe. Your answers should not include any harmful, unethical, racist, sexist, toxic, dangerous, or illegal content. Please ensure that your responses are socially unbiased and positive in nature.\n\nIf a question does not make any sense, or is not factually coherent, explain why instead of answering something not correct. If you don't know the answer to a question, please don't share false information."

	if messages[0].Role == openai.ChatMessageRoleSystem {
		system = strings.TrimSpace(messages[0].Content)
		messages = messages[1:]
	}

	var prompt string

	for i, message := range messages {
		if message.Role == openai.ChatMessageRoleUser {
			content := strings.TrimSpace(message.Content)

			if i == 0 && len(system) > 0 {
				content = "<<SYS>>\n" + system + "\n<</SYS>>\n\n" + content
			}

			prompt += " [INST] " + content + " [/INST]"
		}

		if message.Role == openai.ChatMessageRoleAssistant {
			content := strings.TrimSpace(message.Content)
			prompt += " " + content
		}
	}

	result := &completionRequest{
		Stream: req.Stream,

		Prompt: strings.TrimSpace(prompt),
		Stop:   []string{"[INST]"},

		Temperature: req.Temperature,
		TopP:        req.TopP,

		NPredict: -1,
		//NPredict: 400,
	}

	if result.TopP == 0 {
		result.TopP = 0.9
		//result.TopP = 0.95
	}

	if result.Temperature == 0 {
		result.Temperature = 0.6
		//result.Temperature = 0.2
	}

	return result, nil
}

type completionRequest struct {
	Stream bool `json:"stream"`

	Stop   []string `json:"stop"`
	Prompt string   `json:"prompt"`

	Temperature float32 `json:"temperature"`
	NPredict    int     `json:"n_predict"`
	TopP        float32 `json:"top_p"`
}

type completionResponse struct {
	Model string `json:"model"`

	Stop    bool   `json:"stop"`
	Prompt  string `json:"prompt"`
	Content string `json:"content"`

	Truncated bool `json:"truncated"`
}
