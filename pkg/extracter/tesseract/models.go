package tesseract

type Result struct {
	Data struct {
		// Exit struct {
		// 	Code   int `json:"code"`
		// 	Signal any `json:"signal"`
		// } `json:"exit"`

		Stderr string `json:"stderr"`
		Stdout string `json:"stdout"`
	} `json:"data"`
}
