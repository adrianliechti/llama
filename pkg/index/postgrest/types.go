package postgrest

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Document struct {
	ID string `json:"id"`

	Title   string `json:"title"`
	Source  string `json:"source"`
	Content string `json:"content"`

	Embedding []float32 `json:"embedding"`
}

func (d *Document) UnmarshalJSON(data []byte) error {
	var alias struct {
		ID string `json:"id"`

		Title   string `json:"title"`
		Source  string `json:"source"`
		Content string `json:"content"`

		Embedding string `json:"embedding"`
	}

	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	d.ID = alias.ID
	d.Title = alias.Title
	d.Source = alias.Source

	d.Content = alias.Content

	slices := strings.Split(strings.Trim(alias.Embedding, "[]"), ",")

	for _, slice := range slices {
		var value float32

		if _, err := fmt.Sscanf(slice, "%f", &value); err != nil {
			return err
		}

		d.Embedding = append(d.Embedding, value)
	}

	return nil
}
