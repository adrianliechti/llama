package api

type Document struct {
	ID string `json:"id,omitempty"`

	Content string `json:"content,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`
}

type Result struct {
	Score    *float32 `json:"score,omitempty"`
	Document `json:",inline"`
}

type Query struct {
	Text string `json:"text,omitempty"`

	Limit *int `json:"limit,omitempty"`
}
