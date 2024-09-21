package retriever

type Result struct {
	Title    string `json:"title,omitempty"`
	Content  string `json:"content,omitempty"`
	Location string `json:"location,omitempty"`
}
