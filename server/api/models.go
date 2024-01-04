package api

type Document struct {
	ID string `json:"id"`

	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

type SearchRequest struct {
	Content   string    `json:"content,omitempty"`
	Embedding []float32 `json:"embedding,omitempty"`

	TopK *int     `json:"top_k,omitempty"`
	TopP *float32 `json:"top_p,omitempty"`
}
