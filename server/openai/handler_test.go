package openai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type TestContext struct {
	Context context.Context
	Client  *openai.Client

	Model     string
	Embedding openai.EmbeddingModel
}

func newTestContext() *TestContext {
	config := openai.DefaultConfig("")
	config.BaseURL = "http://localhost:8080/v1"

	client := openai.NewClientWithConfig(config)

	return &TestContext{
		Context: context.Background(),
		Client:  client,

		Model:     openai.GPT3Dot5Turbo,
		Embedding: openai.AdaEmbeddingV2,
	}
}

func TestModels(t *testing.T) {
	c := newTestContext()

	resp, err := c.Client.ListModels(c.Context)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Models)

	for _, model := range resp.Models {
		assert.NotEmpty(t, model.ID)
		assert.NotEmpty(t, model.CreatedAt)
		assert.Equal(t, "model", model.Object)
	}

}

func TestEmbedding(t *testing.T) {
	c := newTestContext()

	resp, err := c.Client.CreateEmbeddings(c.Context, &openai.EmbeddingRequest{
		Model: c.Embedding,
		Input: "The food was delicious and the waiter...",

		EncodingFormat: openai.EmbeddingEncodingFormatFloat,
	})

	assert.NoError(t, err)
	assert.Equal(t, "list", resp.Object)
	assert.NotEmpty(t, resp.Model)
	assert.Len(t, resp.Data, 1)

	if len(resp.Data) == 0 {
		t.FailNow()
	}

	embedding := resp.Data[0]
	assert.Equal(t, embedding.Object, "embedding")
	assert.Equal(t, 0, embedding.Index)
	assert.NotEmpty(t, embedding.Embedding)
}

func TestChatCompletion(t *testing.T) {
	c := newTestContext()

	resp, err := c.Client.CreateChatCompletion(c.Context, openai.ChatCompletionRequest{
		Model: c.Model,

		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful assistant.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Who won the world series in 2020?",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "The Los Angeles Dodgers won the World Series in 2020.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Where was it played?",
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.NotEmpty(t, resp.ID)
	assert.NotEmpty(t, resp.Model)
	assert.NotEmpty(t, resp.Created)
	assert.Len(t, resp.Choices, 1)

	if len(resp.Choices) == 0 {
		t.FailNow()
	}

	choice := resp.Choices[0]
	assert.Equal(t, 0, choice.Index)
	assert.Equal(t, openai.FinishReasonStop, choice.FinishReason)

	assert.Equal(t, openai.ChatMessageRoleAssistant, choice.Message.Role)
	assert.NotEmpty(t, choice.Message.Content)
}

func TestChatCompletionStream(t *testing.T) {
	c := newTestContext()

	stream, err := c.Client.CreateChatCompletionStream(c.Context, openai.ChatCompletionRequest{
		Model: c.Model,

		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful assistant.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Who won the world series in 2020?",
			},
			{
				Role:    openai.ChatMessageRoleAssistant,
				Content: "The Los Angeles Dodgers won the World Series in 2020.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Where was it played?",
			},
		},
	})

	assert.NoError(t, err)

	defer stream.Close()

	var chunks []openai.ChatCompletionStreamResponse

	for {
		resp, err := stream.Recv()

		if err != nil {
			break
		}

		chunks = append(chunks, resp)
	}

	var content strings.Builder

	for i, chunk := range chunks {
		assert.NoError(t, err)
		assert.Equal(t, "chat.completion.chunk", chunk.Object)
		assert.NotEmpty(t, chunk.ID)
		assert.NotEmpty(t, chunk.Created)
		assert.NotEmpty(t, chunk.Model)
		assert.Len(t, chunk.Choices, 1)

		if len(chunk.Choices) == 0 {
			t.FailNow()
		}

		choice := chunk.Choices[0]
		assert.Equal(t, 0, choice.Index)

		if i == len(chunks)-1 {
			assert.Equal(t, openai.FinishReasonStop, choice.FinishReason)
		} else {
			assert.Equal(t, openai.FinishReason(""), choice.FinishReason)
		}

		content.WriteString(choice.Delta.Content)
	}

	assert.NotEmpty(t, content.String())
}

func TestChatCompletionWithTool(t *testing.T) {
	c := newTestContext()

	fnGetCurrentWeather := openai.Tool{
		Type: openai.ToolTypeFunction,

		Function: &openai.FunctionDefinition{
			Description: "Get the current weather in a given location (like a city or region).",

			Name: "get_current_weather",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,

				Properties: map[string]jsonschema.Definition{
					"query": {
						Type:        jsonschema.String,
						Description: "The city and state, e.g. San Francisco, CA",
					},

					"unit": {
						Type: jsonschema.String,
						Enum: []string{"celsius", "fahrenheit"},
					},
				},

				Required: []string{"location"},
			},
		},
	}

	dialoge := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a helpful assistant.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Who won the world series in 2020?",
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "The Los Angeles Dodgers won the World Series in 2020.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "What is the weather in Boston today?",
		},
	}

	resp, err := c.Client.CreateChatCompletion(c.Context, openai.ChatCompletionRequest{
		Model: c.Model,

		Tools: []openai.Tool{
			fnGetCurrentWeather,
		},

		Messages: dialoge,
	})

	assert.NoError(t, err)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.NotEmpty(t, resp.ID)
	assert.NotEmpty(t, resp.Model)
	assert.NotEmpty(t, resp.Created)
	assert.Len(t, resp.Choices, 1)

	if len(resp.Choices) == 0 {
		t.FailNow()
	}

	choice := resp.Choices[0]
	assert.Equal(t, 0, choice.Index)
	assert.Equal(t, openai.FinishReasonToolCalls, choice.FinishReason)
	assert.Len(t, choice.Message.ToolCalls, 1)

	if len(choice.Message.ToolCalls) == 0 {
		t.FailNow()
	}

	dialoge = append(dialoge, choice.Message)

	call := choice.Message.ToolCalls[0]
	assert.NotEmpty(t, call.ID)

	type argsGetCurrentWeather struct {
		Query string `json:"query"`
		Unit  string `json:"unit"`
	}

	var args argsGetCurrentWeather
	assert.NoError(t, json.Unmarshal([]byte(call.Function.Arguments), &args))

	assert.NotEmpty(t, args.Query)

	dialoge = append(dialoge, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleTool,
		Content: "Sunny and 80 degrees.",

		// Name:       msg.ToolCalls[0].Function.Name,
		ToolCallID: call.ID,
	})

	resp, err = c.Client.CreateChatCompletion(c.Context, openai.ChatCompletionRequest{
		Model: c.Model,

		Tools: []openai.Tool{
			fnGetCurrentWeather,
		},

		Messages: dialoge,
	})

	assert.NoError(t, err)
	assert.Equal(t, "chat.completion", resp.Object)
	assert.NotEmpty(t, resp.ID)
	assert.NotEmpty(t, resp.Model)
	assert.NotEmpty(t, resp.Created)
	assert.Len(t, resp.Choices, 1)

	if len(resp.Choices) == 0 {
		t.FailNow()
	}

	choice = resp.Choices[0]
	assert.Equal(t, 0, choice.Index)
	assert.Equal(t, openai.FinishReasonStop, choice.FinishReason)

	assert.Equal(t, openai.ChatMessageRoleAssistant, choice.Message.Role)
	assert.NotEmpty(t, choice.Message.Content)

	dialoge = append(dialoge, choice.Message)

	output, _ := json.MarshalIndent(dialoge, "", "  ")
	t.Log(string(output))
}
