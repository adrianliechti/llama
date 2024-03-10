package api

type Document struct {
	ID string `json:"id,omitempty"`

	Content string `json:"content,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`
}

type Result struct {
	Document `json:",inline"`
	Distance *float32 `json:"distance,omitempty"`
}

type Query struct {
	Text string `json:"text,omitempty"`

	Limit    *int     `json:"limit,omitempty"`
	Distance *float32 `json:"distance,omitempty"`
}
