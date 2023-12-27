package models

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input any    `json:"input"`
}

type Embeddings struct {
	Object string `json:"object"`

	Model string      `json:"model"`
	Data  []Embedding `json:"data"`
}

type Embedding struct {
	Object string `json:"object"`

	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}
