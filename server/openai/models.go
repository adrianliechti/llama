package openai

import (
	"encoding/json"
	"errors"
)

// https://platform.openai.com/docs/api-reference/models/object
type Model struct {
	Object string `json:"object"` // "model"

	ID      string `json:"id"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// https://platform.openai.com/docs/api-reference/models
type ModelList struct {
	Object string `json:"object"` // "list"

	Models []Model `json:"data"`
}

// https://platform.openai.com/docs/api-reference/embeddings/create
type EmbeddingsRequest struct {
	Input any    `json:"input"`
	Model string `json:"model"`

	// encoding_format string: float, base64
	// dimensions int
	// user string
}

// https://platform.openai.com/docs/api-reference/embeddings/object
type Embedding struct {
	Object string `json:"object"` // "embedding"

	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// https://platform.openai.com/docs/api-reference/embeddings/create
type EmbeddingList struct {
	Object string `json:"object"` // "list"

	Model string      `json:"model"`
	Data  []Embedding `json:"data"`

	Usage *Usage `json:"usage,omitempty"`
}

type MessageRole string

var (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
)

type ResponseFormat string

var (
	ResponseFormatText ResponseFormat = "text"
	ResponseFormatJSON ResponseFormat = "json_object"
)

// https://platform.openai.com/docs/api-reference/chat/object
type FinishReason string

var (
	FinishReasonStop   FinishReason = "stop"
	FinishReasonLength FinishReason = "length"

	FinishReasonToolCalls     FinishReason = "tool_calls"
	FinishReasonContentFilter FinishReason = "content_filter"
)

// https://platform.openai.com/docs/api-reference/chat/create
type ChatCompletionRequest struct {
	Model string `json:"model"`

	Messages []ChatCompletionMessage `json:"messages"`

	Stream bool   `json:"stream,omitempty"`
	Stop   any    `json:"stop,omitempty"`
	Tools  []Tool `json:"tools,omitempty"`

	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`

	ResponseFormat *ChatCompletionResponseFormat `json:"response_format,omitempty"`

	// frequency_penalty *float32
	// presence_penalty *float32

	// logit_bias
	// logprobs *bool
	// top_logprobs *int

	// n *int

	// seed *int

	// top_p *float32

	// tool_choice string: none, auto

	// user string
}

// https://platform.openai.com/docs/api-reference/chat/create
type ChatCompletionResponseFormat struct {
	Type ResponseFormat `json:"type"`
}

// https://platform.openai.com/docs/api-reference/chat/object
type ChatCompletion struct {
	Object string `json:"object"` // "chat.completion" | "chat.completion.chunk"

	ID string `json:"id"`

	Model   string `json:"model"`
	Created int64  `json:"created"`

	Choices []ChatCompletionChoice `json:"choices"`

	Usage *Usage `json:"usage,omitempty"`
}

// https://platform.openai.com/docs/api-reference/chat/object
type ChatCompletionChoice struct {
	Index int `json:"index"`

	Delta   *ChatCompletionMessage `json:"delta,omitempty"`
	Message *ChatCompletionMessage `json:"message,omitempty"`

	FinishReason *FinishReason `json:"finish_reason"`
}

// https://platform.openai.com/docs/api-reference/chat/object
type ChatCompletionMessage struct {
	Role MessageRole `json:"role,omitempty"`

	Content  string           `json:"content"`
	Contents []MessageContent `json:"-"`

	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type MessageContent struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`

	ImageURL *MessageContentURL `json:"image_url,omitempty"`
}

type MessageContentURL struct {
	URL string `json:"url"`
}

func (m *ChatCompletionMessage) MarshalJSON() ([]byte, error) {
	if m.Content != "" && m.Contents != nil {
		return nil, errors.New("cannot have both content and contents")
	}

	if len(m.Contents) > 0 {
		type2 := struct {
			Role MessageRole `json:"role"`

			Content  string           `json:"-"`
			Contents []MessageContent `json:"content,omitempty"`

			ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
			ToolCallID string     `json:"tool_call_id,omitempty"`
		}(*m)

		return json.Marshal(type2)
	} else {
		type1 := struct {
			Role MessageRole `json:"role"`

			Content  string           `json:"content"`
			Contents []MessageContent `json:"-"`

			ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
			ToolCallID string     `json:"tool_call_id,omitempty"`
		}(*m)

		return json.Marshal(type1)
	}
}

func (m *ChatCompletionMessage) UnmarshalJSON(data []byte) error {
	type1 := struct {
		Role MessageRole `json:"role"`

		Content  string `json:"content"`
		Contents []MessageContent

		ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
		ToolCallID string     `json:"tool_call_id,omitempty"`
	}{}

	if err := json.Unmarshal(data, &type1); err == nil {
		*m = ChatCompletionMessage(type1)
		return nil
	}

	type2 := struct {
		Role MessageRole `json:"role"`

		Content  string
		Contents []MessageContent `json:"content"`

		ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
		ToolCallID string     `json:"tool_call_id,omitempty"`
	}{}

	if err := json.Unmarshal(data, &type2); err == nil {
		*m = ChatCompletionMessage(type2)
		return err
	}

	return nil
}

// https://platform.openai.com/docs/api-reference/chat/object
type ToolType string

var (
	ToolTypeFunction ToolType = "function"
)

type Tool struct {
	Type ToolType `json:"type"`

	ToolFunction *Function `json:"function"`
}

// https://platform.openai.com/docs/api-reference/chat/object
type ToolCall struct {
	ID string `json:"id"`

	Type ToolType `json:"type"`

	Index int `json:"index"`

	Function *FunctionCall `json:"function,omitempty"`
}

type Function struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Parameters map[string]any `json:"parameters"`
}

// https://platform.openai.com/docs/api-reference/chat/object
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// https://platform.openai.com/docs/api-reference/audio/createSpeech
type SpeechRequest struct {
	Model string `json:"model"`

	Input string `json:"input"`
	Voice string `json:"voice"`
}

type Transcription struct {
	Task string `json:"task"`

	Language string  `json:"language"`
	Duration float64 `json:"duration"`

	Text string `json:"text"`
}

// https://platform.openai.com/docs/api-reference/images/create
type ImageCreateRequest struct {
	Model string `json:"model"`

	Prompt string `json:"prompt"`
	Style  string `json:"style,omitempty"`

	ResponseFormat string `json:"response_format,omitempty"`
}

// https://platform.openai.com/docs/api-reference/images/create
type ImageList struct {
	Images []Image `json:"data"`
}

// https://platform.openai.com/docs/api-reference/images/object
type Image struct {
	URL     string `json:"url,omitempty"`
	B64JSON string `json:"b64_json,omitempty"`

	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type ErrorResponse struct {
	Error Error `json:"error,omitempty"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
