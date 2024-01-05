package api

type Document struct {
	ID string `json:"id"`

	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Result struct {
	Document `json:",inline"`
	Distance float32 `json:"distance"`
}

type Query struct {
	Text      string    `json:"text,omitempty"`
	Embedding []float32 `json:"embedding,omitempty"`

	Limit    *int     `json:"limit,omitempty"`
	Distance *float32 `json:"distance,omitempty"`
}
