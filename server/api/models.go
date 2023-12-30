package api

type Document struct {
	ID string `json:"id"`

	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}
