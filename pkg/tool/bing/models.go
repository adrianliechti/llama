package bing

type Result struct {
	Title    string
	Content  string
	Location string
}

type SearchResponse struct {
	WebPages struct {
		Value []WebPage `json:"value"`
	} `json:"webPages"`
}

type WebPage struct {
	ID  string `json:"id"`
	URL string `json:"url"`

	Name    string `json:"name"`
	Snippet string `json:"snippet"`
}
