package bing

type response struct {
	WebPages struct {
		Value []page `json:"value"`
	} `json:"webPages"`
}

type page struct {
	ID  string `json:"id"`
	URL string `json:"url"`

	Name    string `json:"name"`
	Snippet string `json:"snippet"`
}
