package sbert

type VectorsRequest struct {
	Text string `json:"text"`
}

type VectorsResponse struct {
	Text   string    `json:"text"`
	Vector []float32 `json:"vector"`
}
