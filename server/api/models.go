package api

type Result struct {
	Index    int     `json:"index,omitempty"`
	Score    float64 `json:"score,omitempty"`
	Document `json:",inline"`
}

type Document struct {
	Content string `json:"content,omitempty"`
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
	Content string `json:"content"`

	SegmentLength  *int `json:"segment_length"`
	SegmentOverlap *int `json:"segment_overlap"`
}

type SegmentResponse struct {
	Segements []Segment `json:"segements"`
}

type Segment struct {
	Text string `json:"text"`
}

type SummarizeRequest struct {
	Model string `json:"model"`

	Content string `json:"content"`
}
