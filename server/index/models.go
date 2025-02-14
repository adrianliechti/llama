package index

type Page[T any] struct {
	Items  []T    `json:"items,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

type Document struct {
	ID string `json:"id,omitempty"`

	Title   string `json:"title,omitempty"`
	Source  string `json:"source,omitempty"`
	Content string `json:"content,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`

	Embedding []float32 `json:"embedding,omitempty"`
}

type Result struct {
	Score    *float64 `json:"score,omitempty"`
	Document `json:",inline"`
}

type Query struct {
	Text string `json:"text,omitempty"`

	Limit *int `json:"limit,omitempty"`
}
