package ollama

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

type MessageRole string

var (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
}

type ChatRequest struct {
	Model string `json:"model"`

	Stream *bool  `json:"stream,omitempty"`
	Format string `json:"format,omitempty"`

	Messages []Message `json:"messages"`
}

type ChatResponse struct {
	Model   string  `json:"model"`
	Message Message `json:"message"`

	Done bool `json:"done"`
}
