package custom

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/adrianliechti/llama/pkg/provider"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	_ provider.Completer = (*Provider)(nil)
)

type Provider struct {
	url string

	model string

	client CompleterClient
}

type Option func(*Provider)

func New(url string, options ...Option) (*Provider, error) {
	p := &Provider{
		url: url,
	}

	for _, option := range options {
		option(p)
	}

	if p.url == "" || !strings.HasPrefix(p.url, "grpc://") {
		return nil, errors.New("invalid url")
	}

	url = strings.TrimPrefix(p.url, "grpc://")

	conn, err := grpc.Dial(url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	p.client = NewCompleterClient(conn)

	return p, nil
}

func WithModel(model string) Option {
	return func(p *Provider) {
		p.model = model
	}
}

func (p *Provider) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = &provider.CompleteOptions{}
	}

	if options.Stream != nil {
		defer close(options.Stream)
	}

	stream, err := p.client.Complete(ctx, &CompletionRequest{
		Model: p.model,

		Messages: fromMessages(messages),

		Temperature: options.Temperature,
		TopP:        options.TopP,
		MinP:        options.MinP,
	})

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
