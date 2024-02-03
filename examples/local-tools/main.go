package main

import (
	"bufio"
	"context"
	"os"
	"strings"

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
		},
	}

	fnBookFlight := openai.Tool{
		Type: openai.ToolTypeFunction,

		Function: openai.FunctionDefinition{
			Name:        "book_flight",
			Description: "Use this tool to make a reservation for a flight",
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
						Role:    openai.ChatMessageRoleTool,
						Content: "14:00 ZRHLON14, 18:00 ZRHLON18, 21:00 ZRHLON21",

						ToolCallID: c.ID,
					}

				} else if strings.Contains(c.Function.Name, "book") {
					message = &openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleTool,
						Content: "Flight is booked!",
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

		result := completion.Choices[0].Message

		if err != nil {
			panic(err)
		}

		if result.Role == openai.ChatMessageRoleAssistant && len(result.Content) > 0 {
			output.WriteString(strings.TrimSpace(result.Content))
			output.WriteString("\n")
		}

		messages = append(messages, result)
	}
}
