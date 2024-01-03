package llama

type EmbeddingRequest struct {
	Content string `json:"content"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type CompletionRequest struct {
	Prompt string `json:"prompt"`

	Stream  bool   `json:"stream,omitempty"`
	Grammar string `json:"grammar,omitempty"`

	Temperature *float32 `json:"temperature,omitempty"`
	TopP        *float32 `json:"top_p,omitempty"`
	MinP        *float32 `json:"min_p,omitempty"`

	Stop []string `json:"stop,omitempty"`

	CachePrompt bool `json:"cache_prompt,omitempty"`
}

type CompletionResponse struct {
	Model string `json:"model"`

	Prompt  string `json:"prompt"`
	Content string `json:"content"`

	Stop      bool `json:"stop"`
	Truncated bool `json:"truncated"`
}
