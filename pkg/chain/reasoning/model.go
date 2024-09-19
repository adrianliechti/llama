package reasoning

type Step struct {
	Title   string `json:"title"`
	Content string `json:"content"`

	NextAction Action `json:"next_action"`
}

type Action string

const (
	ActionContinue    = "continue"
	ActionFinalAnswer = "final_answer"
)
