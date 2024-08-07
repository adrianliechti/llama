package tavily

type Result struct {
	Title    string
	Content  string
	Location string
}

type SearchResult struct {
	Query string `json:"query"`

	Answer string `json:"answer"`

	Results []struct {
		URL string `json:"url"`

		Title   string `json:"title"`
		Content string `json:"content"`

		Score float64 `json:"score"`
	} `json:"results"`
}
