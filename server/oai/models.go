package oai

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
)

type ChatCompletionRequest struct {
	Model string `json:"model"`

	Stream bool `json:"stream,omitempty"`

	Format *ChatCompletionResponseFormat `json:"response_format,omitempty"`

	Temperature *float32 `json:"temperature,omitempty"`
	TopP        *float32 `json:"top_p,omitempty"`

	Messages []ChatCompletionMessage `json:"messages"`
}

type ChatCompletionResponseFormat struct {
	Type ResponseFormat `json:"type"`
}

type ChatCompletion struct {
	Object string `json:"object"` // "chat.completion"

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
	Role    MessageRole `json:"role,omitempty"`
	Content string      `json:"content"`
}
