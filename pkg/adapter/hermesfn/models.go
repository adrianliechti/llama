package hermesfn

import (
	"bytes"
	"encoding/json"
)

type Tool struct {
	Type string `json:"type"`

	Function *ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Parameters any `json:"parameters"`
}

type ToolCallback struct {
	Name    string          `json:"name"`
	Content json.RawMessage `json:"content"`
}

type ToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func encodeJSON(v any) (string, error) {
	var data bytes.Buffer

	enc := json.NewEncoder(&data)
	enc.SetEscapeHTML(false)
	enc.Encode(v)

	return data.String(), nil
}
