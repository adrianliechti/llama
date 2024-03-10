package langchain

type MessageType string

var (
	MessageTypeSystem MessageType = "system"
	MessageTypeHuman  MessageType = "human"
	MessageTypeAI     MessageType = "ai"
)

type Message struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
}

type Input struct {
	Input   string    `json:"input"`
	History []Message `json:"chat_history"`
}

type RunInput struct {
	Input Input `json:"input"`
}

type RunData struct {
	Output  string `json:"output"`
	Content string `json:"content"`

	Messages []Message `json:"messages"`
}
