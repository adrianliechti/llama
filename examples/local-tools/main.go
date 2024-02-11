package main

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/adrianliechti/llama/pkg/jsonschema"
	"github.com/sashabaranov/go-openai"
)

func main() {
	config := openai.DefaultConfig("")
	config.BaseURL = "http://localhost:8080/oai/v1"

	model := "mistral-fn"

	client := openai.NewClientWithConfig(config)

	fnFindFlight := openai.Tool{
		Type: openai.ToolTypeFunction,

		Function: openai.FunctionDefinition{
			Name:        "find_flight",
			Description: "Use this tool to find flight numbers form a city to another city",

			Parameters: jsonschema.Definition{
				Type: jsonschema.DataTypeObject,

				Properties: map[string]jsonschema.Definition{
					"departure": {
						Type:        jsonschema.DataTypeString,
						Description: "Departure Airport Code or Name",
					},

					"arrival": {
						Type:        jsonschema.DataTypeString,
						Description: "Arrival Airport Code or Name",
					},
				},

				Required: []string{"departure", "arrival"},
			},
		},
	}

	fnBookFlight := openai.Tool{
		Type: openai.ToolTypeFunction,

		Function: openai.FunctionDefinition{
			Name:        "book_flight",
			Description: "Use this tool to make a reservation for a flight",

			Parameters: jsonschema.Definition{
				Type: jsonschema.DataTypeObject,

				Properties: map[string]jsonschema.Definition{
					"departure": {
						Type:        jsonschema.DataTypeString,
						Description: "Departure Airport Code or Name",
					},

					"arrival": {
						Type:        jsonschema.DataTypeString,
						Description: "Arrival Airport Code or Name",
					},

					"time": {
						Type:        jsonschema.DataTypeString,
						Description: "Desired Departure Time",
					},
				},

				Required: []string{"departure", "arrival", "time"},
			},
		},
	}

	ctx := context.Background()

	reader := bufio.NewReader(os.Stdin)
	output := os.Stdout

	var messages []openai.ChatCompletionMessage

	var message *openai.ChatCompletionMessage

	for {
		message = nil

		if len(messages) > 0 && len(messages[len(messages)-1].ToolCalls) > 0 {
			println("tool call")

			calls := messages[len(messages)-1].ToolCalls

			for _, c := range calls {
				println("fn", c.Function.Name, c.Function.Arguments)

				if strings.Contains(c.Function.Name, "find") {
					message = &openai.ChatCompletionMessage{
						Role: openai.ChatMessageRoleTool,

						ToolCallID: c.ID,
						Content:    `[ { "id": "ZRHLON14", "time": "14:00" }, { "id": "ZRHLON18", "time": "18:00" }, { "id": "ZRHLON21", "time": "21:00" } ]`,
					}

				} else if strings.Contains(c.Function.Name, "book") {
					message = &openai.ChatCompletionMessage{
						Role: openai.ChatMessageRoleTool,

						ToolCallID: c.ID,
						Content:    `{ "status": "booked" }`,
					}
				}
			}
		}

		if message == nil {
			output.WriteString(">>> ")
			input, err := reader.ReadString('\n')

			if err != nil {
				panic(err)
			}

			message = &openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: strings.TrimSpace(input),
			}
		}

		messages = append(messages, *message)

		req := openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,

			Tools: []openai.Tool{
				fnFindFlight,
				fnBookFlight,
			},
		}

		completion, err := client.CreateChatCompletion(ctx, req)

		if err != nil {
			panic(err)
		}

		result := completion.Choices[0].Message

		if result.Role == openai.ChatMessageRoleAssistant && len(result.Content) > 0 {
			output.WriteString(strings.TrimSpace(result.Content))
			output.WriteString("\n")
		}

		messages = append(messages, result)
	}
}
