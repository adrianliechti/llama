package unstructured

type Element struct {
	ID string `json:"element_id"`

	Type string `json:"type"`
	Text string `json:"text"`

	Metadata ElementMetadata `json:"metadata"`
}

type ElementMetadata struct {
	Languages  []string `json:"languages"`
	PageNumber int      `json:"page_number"`
	Filename   string   `json:"filename"`
	Filetype   string   `json:"filetype"`
}
