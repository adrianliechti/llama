package unstructured

type Strategy string

const (
	StrategyAuto  Strategy = "auto"
	StrategyFast  Strategy = "fast"
	StrategyHiRes Strategy = "hi_res"
)

type Element struct {
	ID string `json:"element_id"`

	Type string `json:"type"`
	Text string `json:"text"`

	Metadata ElementMetadata `json:"metadata"`
}

type ElementMetadata struct {
	FileName string `json:"filename"`
	FileType string `json:"filetype"`

	Languages []string `json:"languages"`

	// PageName   string `json:"page_name"`
	// PageNumber int    `json:"page_number"`

	// MailSender    string `json:"sent_from"`
	// MailRecipient string `json:"sent_to"`
	// MailSubject   string `json:"subject"`
}
