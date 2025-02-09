package searxng

type searchResponse struct {
	Results []result `json:"results"`
}

type result struct {
	URL string `json:"url"`

	Engine string `json:"engine"`

	Title   string `json:"title"`
	Content string `json:"content"`

	Score float32 `json:"score"`
}
