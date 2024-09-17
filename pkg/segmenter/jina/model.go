package jina

type SegmentRequest struct {
	Content string `json:"content"`

	ReturnChunks bool `json:"return_chunks,omitempty"`

	MaxChunkLength int `json:"max_chunk_length,omitempty"`
}

type SegmentResponse struct {
	Chunks []string `json:"chunks"`
}
