package google

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"unicode"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/to"
)

var _ provider.Completer = (*Completer)(nil)

type Completer struct {
	*Config
}

func NewCompleter(model string, options ...Option) (*Completer, error) {
	cfg := &Config{
		model: model,
	}

	for _, option := range options {
		option(cfg)
	}

	return &Completer{
		Config: cfg,
	}, nil
}

func (c *Completer) Complete(ctx context.Context, messages []provider.Message, options *provider.CompleteOptions) (*provider.Completion, error) {
	if options == nil {
		options = new(provider.CompleteOptions)
	}

	client, err := genai.NewClient(ctx, c.Options()...)

	if err != nil {
		return nil, err
	}

	defer client.Close()

	system, err := convertSystem(messages)

	if err != nil {
		return nil, err
	}

	history, err := convertHistory(messages)

	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel(c.model)
	model.SystemInstruction = system

	if len(options.Tools) > 0 {
		model.Tools = convertTools(options.Tools)
	}

	if len(options.Stop) > 0 {
		model.StopSequences = options.Stop
	}

	if options.MaxTokens != nil {
		model.SetMaxOutputTokens(int32(*options.MaxTokens))
	}

	if options.Temperature != nil {
		model.SetTemperature(*options.Temperature)
	}

	if options.Format == provider.CompletionFormatJSON || options.Schema != nil {
		model.ResponseMIMEType = "application/json"

		if options.Schema != nil {
			model.ResponseSchema = convertSchema(options.Schema.Schema)
		}
	}

	session := model.StartChat()
	session.History = history

	prompt, err := convertContent(messages[len(messages)-1])

	if err != nil {
		return nil, err
	}

	if options.Stream != nil {
		return c.completeStream(ctx, session, prompt.Parts, options)
	}

	return c.complete(ctx, session, prompt.Parts, options)
}

func (c *Completer) complete(ctx context.Context, session *genai.ChatSession, parts []genai.Part, options *provider.CompleteOptions) (*provider.Completion, error) {
	resp, err := session.SendMessage(ctx, parts...)

	if err != nil {
		return nil, convertError(err)
	}

	candidate := resp.Candidates[0]

	content := toContent(candidate.Content)
	content = strings.TrimRightFunc(content, unicode.IsSpace)

	return &provider.Completion{
		ID:     uuid.New().String(),
		Reason: toCompletionResult(candidate.FinishReason),

		Message: provider.Message{
			Role:    provider.MessageRoleAssistant,
			Content: content,

			ToolCalls: toToolCalls(candidate.Content),
		},
	}, nil
}

func (c *Completer) completeStream(ctx context.Context, session *genai.ChatSession, parts []genai.Part, options *provider.CompleteOptions) (*provider.Completion, error) {
	iter := session.SendMessageStream(ctx, parts...)

	result := &provider.Completion{
		ID: uuid.New().String(),

		Message: provider.Message{
			Role: provider.MessageRoleAssistant,
		},

		//Usage: &provider.Usage{},
	}

	resultToolCalls := map[string]provider.ToolCall{}

	for i := 0; ; i++ {
		resp, err := iter.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, convertError(err)
		}

		candidate := resp.Candidates[0]

		content := toContent(candidate.Content)

		if i == 0 {
			content = strings.TrimLeftFunc(content, unicode.IsSpace)
		}

		result.Message.Content += content

		for _, c := range toToolCalls(candidate.Content) {
			resultToolCalls[c.Name] = c
		}

		if len(content) > 0 {
			delta := provider.Completion{
				ID: result.ID,

				Message: provider.Message{
					Role:    provider.MessageRoleAssistant,
					Content: content,
				},
			}

			if err := options.Stream(ctx, delta); err != nil {
				return nil, err
			}
		}
	}

	result.Message.Content = strings.TrimRightFunc(result.Message.Content, unicode.IsSpace)

	if len(resultToolCalls) > 0 {
		result.Message.ToolCalls = to.Values(resultToolCalls)
	}

	return result, nil
}

func convertSystem(messages []provider.Message) (*genai.Content, error) {
	var parts []genai.Part

	for _, m := range messages {
		if m.Role != provider.MessageRoleSystem {
			continue
		}

		if m.Content != "" {
			parts = append(parts, genai.Text(m.Content))
		}
	}

	if len(parts) == 0 {
		return nil, nil
	}

	return &genai.Content{
		Parts: parts,
	}, nil
}

func convertContent(message provider.Message) (*genai.Content, error) {
	var role string
	var parts []genai.Part

	switch message.Role {
	case provider.MessageRoleUser:
		role = "user"

		if message.Content != "" {
			parts = append(parts, genai.Text(message.Content))
		}

		for _, f := range message.Files {
			if strings.HasPrefix(f.ContentType, "image/") {
				format := strings.Split(f.ContentType, "/")[1]

				data, err := io.ReadAll(f.Content)

				if err != nil {
					return nil, err
				}

				parts = append(parts, genai.ImageData(format, data))
			}
		}

	case provider.MessageRoleAssistant:
		role = "model"

		if message.Content != "" {
			parts = append(parts, genai.Text(message.Content))
		}

		for _, c := range message.ToolCalls {
			var data map[string]any
			json.Unmarshal([]byte(c.Arguments), &data)

			part := genai.FunctionCall{
				Name: c.Name,
				Args: data,
			}

			parts = append(parts, part)
		}

	case provider.MessageRoleTool:
		role = "user"

		var data any
		json.Unmarshal([]byte(message.Content), &data)

		var parameters map[string]any

		if val, ok := data.(map[string]any); ok {
			parameters = val
		}

		if val, ok := data.([]any); ok {
			parameters = map[string]any{"data": val}
		}

		part := genai.FunctionResponse{
			Name:     message.Tool,
			Response: parameters,
		}

		parts = append(parts, part)
	}

	content := &genai.Content{
		Role:  role,
		Parts: parts,
	}

	return content, nil
}

func convertHistory(messages []provider.Message) ([]*genai.Content, error) {
	var result []*genai.Content

	if len(messages) < 1 {
		return result, nil
	}

	for _, m := range messages[:len(messages)-1] {
		if m.Role == provider.MessageRoleUser {
			content, err := convertContent(m)

			if err != nil {
				return nil, err
			}

			result = append(result, content)
		}

		if m.Role == provider.MessageRoleAssistant {
			content, err := convertContent(m)

			if err != nil {
				return nil, err
			}

			result = append(result, content)
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func convertTools(tools []provider.Tool) []*genai.Tool {
	var functions []*genai.FunctionDeclaration

	for _, t := range tools {
		function := &genai.FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,

			Parameters: convertSchema(t.Parameters),
		}

		functions = append(functions, function)
	}

	if len(functions) == 0 {
		return nil
	}

	return []*genai.Tool{
		{
			FunctionDeclarations: functions,
		},
	}
}

func convertSchema(parameters map[string]any) *genai.Schema {
	if len(parameters) == 0 {
		return nil
	}

	schema := &genai.Schema{
		Type: genai.TypeObject,
	}

	if val, ok := parameters["type"].(string); ok {
		switch val {
		case "string":
			schema.Type = genai.TypeString
		case "number":
			schema.Type = genai.TypeNumber
		case "integer":
			schema.Type = genai.TypeInteger
		case "boolean ":
			schema.Type = genai.TypeBoolean
		case "array":
			schema.Type = genai.TypeArray
		case "object":
			schema.Type = genai.TypeObject
		}
	}

	if val, ok := parameters["description"].(string); ok {
		schema.Description = val
	}

	if val, ok := parameters["enum"].([]string); ok {
		schema.Enum = val
	}

	if val, ok := parameters["items"].(map[string]any); ok {
		schema.Items = convertSchema(val)
	}

	if val, ok := parameters["properties"].(map[string]any); ok {
		schema.Properties = make(map[string]*genai.Schema)

		for key, value := range val {
			parameters, ok := value.(map[string]any)

			if ok {
				schema.Properties[key] = convertSchema(parameters)
			}
		}
	}

	if val, ok := parameters["required"].([]string); ok {
		schema.Required = val
	}

	return schema
}

func toContent(content *genai.Content) string {
	if content == nil {
		return ""
	}

	for _, p := range content.Parts {
		switch v := p.(type) {
		case genai.Text:
			return string(v)
		}
	}
	return ""
}

func toToolCalls(content *genai.Content) []provider.ToolCall {
	if content == nil {
		return nil
	}

	var result []provider.ToolCall

	for _, p := range content.Parts {
		switch v := p.(type) {
		case genai.FunctionCall:
			data, _ := json.Marshal(v.Args)

			call := provider.ToolCall{
				ID: uuid.NewString(),

				Name:      v.Name,
				Arguments: string(data),
			}

			result = append(result, call)
		}
	}

	return result
}

func toCompletionResult(val genai.FinishReason) provider.CompletionReason {
	switch val {
	case genai.FinishReasonStop:
		return provider.CompletionReasonStop

	case genai.FinishReasonMaxTokens:
		return provider.CompletionReasonLength

	case genai.FinishReasonSafety:
		return provider.CompletionReasonFilter

	case genai.FinishReasonRecitation:
		return provider.CompletionReasonFilter

	default:
		return ""
	}
}
