package oai

import (
	"encoding/json"
	"errors"
)

type Model struct {
	Object string `json:"object"` // "model"

	ID      string `json:"id"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ModelList struct {
	Object string `json:"object"` // "list"

	Models []Model `json:"data"`
}

type EmbeddingsRequest struct {
	Input any    `json:"input"`
	Model string `json:"model"`
}

type Embedding struct {
	Object string `json:"object"` // "embedding"

	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type EmbeddingList struct {
	Object string `json:"object"` // "list"

	Model string      `json:"model"`
	Data  []Embedding `json:"data"`
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

type CompletionReason string

var (
	CompletionReasonStop   CompletionReason = "stop"
	CompletionReasonLength CompletionReason = "length"

	CompletionReasonToolCalls CompletionReason = "tool_calls"
)

type ChatCompletionRequest struct {
	Model string `json:"model"`

	Messages []ChatCompletionMessage `json:"messages"`
	Tools    []Tool                  `json:"tools,omitempty"`

	Stream bool `json:"stream,omitempty"`

	Format *ChatCompletionResponseFormat `json:"response_format,omitempty"`

	Temperature *float32 `json:"temperature,omitempty"`
	TopP        *float32 `json:"top_p,omitempty"`
}

type ChatCompletionResponseFormat struct {
	Type ResponseFormat `json:"type"`
}

type ChatCompletion struct {
	Object string `json:"object"`

	ID string `json:"id"`

	Model   string `json:"model"`
	Created int64  `json:"created"`

	Choices []ChatCompletionChoice `json:"choices"`
}

type ChatCompletionChoice struct {
	Index int `json:"index"`

	Delta   *ChatCompletionMessage `json:"delta,omitempty"`
	Message *ChatCompletionMessage `json:"message,omitempty"`

	FinishReason *CompletionReason `json:"finish_reason"`
}

type ChatCompletionMessage struct {
	Role MessageRole `json:"role,omitempty"`

	Content  string                  `json:"content"`
	Contents []ChatCompletionContent `json:"-"`

	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ChatCompletionContent struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`

	ImageURL *ChatCompletionFile `json:"image_url,omitempty"`
}

type ChatCompletionFile struct {
	URL string `json:"url"`
}

func (m *ChatCompletionMessage) MarshalJSON() ([]byte, error) {
	if m.Content != "" && m.Contents != nil {
		return nil, errors.New("cannot have both content and contents")
	}

	if len(m.Contents) > 0 {
		msg := struct {
			Role MessageRole `json:"role"`

			Content  string                  `json:"-"`
			Contents []ChatCompletionContent `json:"content,omitempty"`

			ToolCallID string     `json:"tool_call_id,omitempty"`
			ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
		}(*m)

		return json.Marshal(msg)
	}

	msg := struct {
		Role MessageRole `json:"role"`

		Content  string                  `json:"content"`
		Contents []ChatCompletionContent `json:"-"`

		ToolCallID string     `json:"tool_call_id,omitempty"`
		ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	}(*m)

	return json.Marshal(msg)
}

func (m *ChatCompletionMessage) UnmarshalJSON(data []byte) error {
	m1 := struct {
		Role MessageRole `json:"role"`

		Content  string `json:"content"`
		Contents []ChatCompletionContent

		ToolCallID string     `json:"tool_call_id,omitempty"`
		ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	}{}

	if err := json.Unmarshal(data, &m1); err == nil {
		*m = ChatCompletionMessage(m1)
		return nil
	}

	m2 := struct {
		Role MessageRole `json:"role"`

		Content  string
		Contents []ChatCompletionContent `json:"content"`

		ToolCallID string     `json:"tool_call_id,omitempty"`
		ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	}{}

	if err := json.Unmarshal(data, &m2); err == nil {
		*m = ChatCompletionMessage(m2)
		return err
	}

	return nil
}

type ToolType string

var (
	ToolTypeFunction ToolType = "function"
)

type Tool struct {
	Type ToolType `json:"type"`

	ToolFunction *Function `json:"function"`
}

type ToolCall struct {
	ID string `json:"id"`

	Type ToolType `json:"type"`

	//Index *int `json:"index,omitempty"`

	Function *FunctionCall `json:"function,omitempty"`
}

type Function struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Parameters any `json:"parameters"`
}

type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type Transcription struct {
	Text string `json:"text"`
}

type ErrorResponse struct {
	Error Error `json:"error,omitempty"`
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
