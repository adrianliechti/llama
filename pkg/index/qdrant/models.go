package qdrant

type payload struct {
	Title   string `json:"title,omitempty"`
	Source  string `json:"source,omitempty"`
	Content string `json:"content,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`
}

type point struct {
	ID string `json:"id"`

	Vector []float32 `json:"vector"`

	Payload payload `json:"payload"`
}

type result struct {
	ID string `json:"id"`

	Version int     `json:"version"`
	Score   float32 `json:"score"`

	Vector []float32 `json:"vector"`

	Payload payload `json:"payload"`
}

type queryResult struct {
	Result []result `json:"result"`
}

type scrollResult struct {
	Result struct {
		Points []point `json:"points"`

		NextPageOffset string `json:"next_page_offset"`
	} `json:"result"`
}
