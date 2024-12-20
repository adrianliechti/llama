package search

type Result struct {
	URL string `json:"url"`

	Title   string `json:"title"`
	Content string `json:"content"`
}
