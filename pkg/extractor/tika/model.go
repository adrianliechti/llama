package tika

type TikaResponse struct {
	Content string `json:"X-TIKA:content"`

	//ContentType string `json:"Content-Type"`
	//ContentLength string `json:"Content-Length"`
}
