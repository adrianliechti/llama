package azuretranslator

import (
	"context"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/google/uuid"
)

func (c *Client) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	message := messages[len(messages)-1]

	result, err := c.Translate(ctx, message.Content, &provider.TranslateOptions{})

	if err != nil {
		return nil, err
	}

	completion := provider.Completion{
		ID: uuid.New().String(),

		Reason: provider.CompletionReasonStop,

		Message: provider.Message{
			Role:    provider.MessageRoleAssistant,
			Content: result.Content,
		},
	}

	// if options.Stream != nil {
	// 	options.Stream <- completion
	// }

	return &completion, nil
}
