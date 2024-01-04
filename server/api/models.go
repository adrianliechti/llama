package api

type Document struct {
	ID string `json:"id"`

	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

type Query struct {
	Text      string    `json:"text,omitempty"`
	Embedding []float32 `json:"embedding,omitempty"`

	Limit *int `json:"limit,omitempty"`
	//TopP  *float32 `json:"top_p,omitempty"`
}
