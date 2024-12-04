package bedrock

import (
	"context"
	"errors"
	"fmt"

	"github.com/adrianliechti/llama/pkg/provider"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config

	client *bedrockruntime.Client
}

func NewCompleter(model string, options ...Option) (*Completer, error) {
	cfg := &Config{
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	config, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		return nil, err
	}

	client := bedrockruntime.NewFromConfig(config)

	return &Completer{
		Config: cfg,

		client: client,
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	msgs, err := convertMessages(messages)

	if err != nil {
		return nil, err
	}

	if options.Stream == nil {
		resp, err := c.client.Converse(ctx, &bedrockruntime.ConverseInput{
			ModelId: aws.String(c.model),

			Messages: msgs,
		})

		if err != nil {
			return nil, err
		}

		return &provider.Completion{
			ID:     uuid.New().String(),
			Reason: toCompletionResult(resp.StopReason),

			Message: provider.Message{
				Role:    provider.MessageRoleAssistant,
				Content: toContent(resp.Output),
			},
		}, nil
	} else {
		resp, err := c.client.ConverseStream(context.Background(), &bedrockruntime.ConverseStreamInput{
			ModelId: aws.String(c.model),

			Messages: msgs,
		})

		if err != nil {
			return nil, err
		}

		result := &provider.Completion{
			ID: uuid.New().String(),

			Message: provider.Message{
				Role: provider.MessageRoleAssistant,
			},

			//Usage: &provider.Usage{},
		}

		for event := range resp.GetStream().Events() {
			switch v := event.(type) {
			case *types.ConverseStreamOutputMemberMessageStart:
				result.Message.Role = toRole(v.Value.Role)

			//case *types.ConverseStreamOutputMemberContentBlockStart:

			case *types.ConverseStreamOutputMemberContentBlockDelta:
				switch c := v.Value.Delta.(type) {
				case *types.ContentBlockDeltaMemberText:
					content := c.Value

					if len(content) > 0 {
						completion := provider.Completion{
							ID: result.ID,

							Message: provider.Message{
								Role:    provider.MessageRoleAssistant,
								Content: content,
							},
						}

						if err := options.Stream(ctx, completion); err != nil {
							return nil, err
						}
					}

					result.Message.Content += content

				// case *types.ContentBlockDeltaMemberToolUse:

				default:
					fmt.Printf("unknown delta type, %T\n", c)
				}

			case *types.ConverseStreamOutputMemberContentBlockStop:

			case *types.ConverseStreamOutputMemberMessageStop:
				reason := toCompletionResult(v.Value.StopReason)

				if reason != "" {
					result.Reason = reason
				}

			case *types.ConverseStreamOutputMemberMetadata:
				result.Usage = &provider.Usage{
					InputTokens:  int(*v.Value.Usage.InputTokens),
					OutputTokens: int(*v.Value.Usage.OutputTokens),
				}

			case *types.UnknownUnionMember:
				fmt.Println("unknown tag:", v.Tag)

			default:
				fmt.Printf("unknown event type, %T\n", v)
			}
		}

		return result, nil
	}
}

func convertMessages(messages []provider.Message) ([]types.Message, error) {
	var result []types.Message

	for _, m := range messages {
		switch m.Role {

		case provider.MessageRoleUser:
			result = append(result, types.Message{
				Role: types.ConversationRoleUser,

				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: m.Content,
					},
				},
			})

		case provider.MessageRoleAssistant:
			result = append(result, types.Message{
				Role: types.ConversationRoleAssistant,

				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: m.Content,
					},
				},
			})

		default:
			return nil, errors.New("unsupported message role")
		}
	}

	return result, nil
}

func toCompletionResult(val types.StopReason) provider.CompletionReason {
	switch val {
	case types.StopReasonEndTurn:
		return provider.CompletionReasonStop

	default:
		return ""
	}
}

func toRole(val types.ConversationRole) provider.MessageRole {
	switch val {
	case types.ConversationRoleUser:
		return provider.MessageRoleUser

	case types.ConversationRoleAssistant:
		return provider.MessageRoleAssistant

	default:
		return ""
	}
}

func toContent(val types.ConverseOutput) string {
	message, ok := val.(*types.ConverseOutputMemberMessage)

	if !ok {
		return ""
	}

	content, ok := message.Value.Content[0].(*types.ContentBlockMemberText)

	if !ok {
		return ""
	}

	return content.Value
}
