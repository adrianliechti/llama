package deepl

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/google/uuid"
)

func (t *Translator) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	message := messages[len(messages)-1]

	result, err := t.Translate(ctx, message.Content, nil)

	if err != nil {
		return nil, err
	}

	completion := provider.Completion{
		ID:     uuid.New().String(),
		Reason: provider.CompletionReasonStop,

		Message: provider.Message{
			Role:    provider.MessageRoleAssistant,
			Content: result.Content,
		},
	}

	return &completion, nil
}
