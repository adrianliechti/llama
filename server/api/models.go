package api

type Document struct {
	ID string `json:"id,omitempty"`

	Content string `json:"content,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`
}

type Result struct {
	Score    *float32 `json:"score,omitempty"`
	Document `json:",inline"`
}

type Query struct {
	Text string `json:"text,omitempty"`

	Limit *int `json:"limit,omitempty"`
}

type Partition struct {
	ID string `json:"element_id,omitempty"`

	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`

	Metadata PartitionMetadata `json:"metadata,omitempty"`
}

type PartitionMetadata struct {
	FileName string `json:"filename,omitempty"`
	FileType string `json:"filetype,omitempty"`

	Languages []string `json:"languages,omitempty"`
}
