package template

import (
	"slices"

	"github.com/adrianliechti/llama/pkg/provider"
)

func ApplyMessage(message provider.Message, data any) (provider.Message, error) {
	t, err := NewTemplate(message.Content)

	if err != nil {
		return message, err
	}

	content, err := t.Execute(data)

	if err != nil {
		return message, err
	}

	message.Content = content

	return message, nil
}

func ApplyMessages(messages []provider.Message, data any) ([]provider.Message, error) {
	result := slices.Clone(messages)

	for i, m := range result {
		message, err := ApplyMessage(m, data)

		if err != nil {
			return nil, err
		}

		result[i] = message
	}

	return result, nil
}
