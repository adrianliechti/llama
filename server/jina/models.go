package jina

type ReadRequest struct {
	URL string `json:"url"`

	PDF  string `json:"pdf"`
	HTML string `json:"html"`
}

type RerankRequest struct {
	Model string `json:"model"`

	Query     string   `json:"query"`
	Documents []string `json:"documents"`

	TopN *int `json:"top_n"`
}

type RerankResponse struct {
	Model string `json:"model"`

	Results []RerankResult `json:"results"`
}

type RerankResult struct {
	Index int `json:"index"`

	Document Document `json:"document"`

	RelevanceScore float64 `json:"relevance_score"`
}

type Document struct {
	Text string `json:"text"`
}

type SegmentRequest struct {
	Content string `json:"content"`

	MaxChunkLength int `json:"max_chunk_length"`
}

type SegmentResponse struct {
	Chunks []string `json:"chunks"`
}
