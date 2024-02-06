package langchain

type InputType string

var (
	InputTypeSystem InputType = "system"
	InputTypeHuman  InputType = "human"
	InputTypeAI     InputType = "ai"
)

type Input struct {
	Type    InputType `json:"type"`
	Content string    `json:"content"`
}

type RunInput struct {
	Input []Input `json:"input"`
}

type DataType string

var (
	DataTypeAIMessageChunk DataType = "AIMessageChunk"
)

type RunData struct {
	Type    DataType `json:"type"`
	Content string   `json:"content"`
}
