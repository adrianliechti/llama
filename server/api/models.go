package api

type Result struct {
	Index    int     `json:"index,omitempty"`
	Score    float64 `json:"score,omitempty"`
	Document `json:",inline"`
}

type Document struct {
	Content string `json:"content,omitempty"`

	Segments []Segment `json:"segments,omitempty"`
}

type RerankRequest struct {
	Model string `json:"model"`

	Query     string   `json:"query"`
	Documents []string `json:"documents"`

	Limit *int `json:"limit,omitempty"`
}

type RerankResponse struct {
	Model string `json:"model"`

	Results []Result `json:"results"`
}

type SegmentRequest struct {
	Text    string `json:"text"`
	Content string `json:"content"` // Deprecated

	SegmentLength  *int `json:"segment_length"`
	SegmentOverlap *int `json:"segment_overlap"`
}

type Segment struct {
	Text string `json:"text"`
}

type SummarizeRequest struct {
	Model string `json:"model"`

	Content string `json:"content"`
}
