package elasticsearch

type Document struct {
	ID string `json:"id"`

	Title   string `json:"title"`
	Source  string `json:"source"`
	Content string `json:"content"`

	Metadata map[string]string `json:"metadata"`
}

type SearchResult struct {
	Hits SearchHits `json:"hits"`
}

type SearchHits struct {
	Hits []SearchHit `json:"hits"`
}

type SearchHit struct {
	Score    float32  `json:"_score"`
	Document Document `json:"_source"`
}
