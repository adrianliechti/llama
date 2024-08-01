package searxng

type Result struct {
	Title    string
	Content  string
	Location string
}

type SearchResult struct {
	URL string `json:"url"`

	Engine string `json:"engine"`

	Title   string `json:"title"`
	Content string `json:"content"`

	Score float32 `json:"score"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}
