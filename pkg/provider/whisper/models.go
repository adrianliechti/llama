package whisper

type InferenceRequest struct {
	Temperature *float32 `json:"temperature,omitempty"`
}

type InferenceResponse struct {
	Text string `json:"text"`
}
