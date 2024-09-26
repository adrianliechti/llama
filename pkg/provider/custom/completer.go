package custom

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/to"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	_ provider.Completer = (*Completer)(nil)
)

type Completer struct {
	*Config
	client CompleterClient
}

func NewCompleter(url string, options ...Option) (*Completer, error) {
	if url == "" || !strings.HasPrefix(url, "grpc://") {
		return nil, errors.New("invalid url")
	}

	c := &Config{
		url: url,
	}

	for _, option := range options {
		option(c)
	}

	client, err := grpc.NewClient(strings.TrimPrefix(c.url, "grpc://"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	return &Completer{
		Config: c,
		client: NewCompleterClient(client),
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	if options.Stream != nil {
		defer close(options.Stream)
	}

	req := &CompletionRequest{
		Model: c.model,

		Messages: fromMessages(messages),
	}

	if options.Stop != nil {
		req.Stop = options.Stop
	}

	if options.MaxTokens != nil {
		req.MaxTokens = to.Ptr(int32(*options.MaxTokens))
	}

	if options.Temperature != nil {
		req.Temperature = options.Temperature
	}

	stream, err := c.client.Complete(ctx, req)

	if err != nil {
		return nil, err
	}

	var result *provider.Completion

	for {
		resp, err := stream.Recv()

		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		completion := provider.Completion{
			ID: resp.Id,

			Message: provider.Message{
				Role:    toRole(resp.Message.Role),
				Content: resp.Message.Content,
			},
		}

		if options.Stream != nil {
			options.Stream <- completion
		}

		result = &completion
	}

	return result, nil
}

func fromRole(role provider.MessageRole) Role {
	switch role {
	case provider.MessageRoleSystem:
		return Role_ROLE_SYSTEM

	case provider.MessageRoleUser:
		return Role_ROLE_USER

	case provider.MessageRoleAssistant:
		return Role_ROLE_ASSISTANT
	}

	return Role_ROLE_UNSPECIFIED
}

func fromMessages(messages []provider.Message) []*Message {
	result := make([]*Message, 0)

	for _, m := range messages {
		message := &Message{
			Role:    fromRole(m.Role),
			Content: m.Content,
		}

		result = append(result, message)
	}

	return result
}

func toRole(role Role) provider.MessageRole {
	switch role {
	case Role_ROLE_SYSTEM:
		return provider.MessageRoleSystem

	case Role_ROLE_USER:
		return provider.MessageRoleUser

	case Role_ROLE_ASSISTANT:
		return provider.MessageRoleAssistant
	}

	return ""
}
