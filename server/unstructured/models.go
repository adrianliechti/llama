package unstructured

type Partition struct {
	ID string `json:"element_id,omitempty"`

	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`

	//Metadata PartitionMetadata `json:"metadata,omitempty"`
}

// type PartitionMetadata struct {
// 	FileName string `json:"filename,omitempty"`
// 	FileType string `json:"filetype,omitempty"`

// 	Languages []string `json:"languages,omitempty"`
// }
