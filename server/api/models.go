package api

type Result struct {
	Index    int     `json:"index,omitempty"`
	Score    float64 `json:"score,omitempty"`
	Document `json:",inline"`
}

type Segment struct {
	Text string `json:"text"`
}

type Document struct {
	Content string `json:"content,omitempty"`

	Segments []Segment `json:"segments,omitempty"`
}
