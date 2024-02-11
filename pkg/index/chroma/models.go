package chroma

type collection struct {
	ID string `json:"id,omitempty"`

	Tenant   string `json:"tenant,omitempty"`
	Database string `json:"database,omitempty"`

	Name     string         `json:"name,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type embeddings struct {
	IDs []string `json:"ids"`

	Embeddings [][]float32 `json:"embeddings"`

	Metadatas []map[string]string `json:"metadatas"`
	Documents []string            `json:"documents"`
}

type getResult struct {
	IDs []string `json:"ids"`

	Distances []float32 `json:"distances,omitempty"`

	Embeddings [][]float64 `json:"embeddings"`

	Metadatas []map[string]string `json:"metadatas"`
	Documents []string            `json:"documents"`
}

type queryResult struct {
	IDs [][]string `json:"ids"`

	Distances [][]float32 `json:"distances,omitempty"`

	Embeddings [][][]float64 `json:"embeddings"`

	Metadatas [][]map[string]string `json:"metadatas"`
	Documents [][]string            `json:"documents"`
}

type errorDetail struct {
	Message string `json:"msg"`
}
