package main

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
	config := openai.DefaultConfig("")
	config.BaseURL = "http://localhost:8080/oai/v1"

	model := "dolphin-llama3"

	client := openai.NewClientWithConfig(config)

	toolGetExchangeRate := openai.Tool{
		Type: openai.ToolTypeFunction,

		Function: &openai.FunctionDefinition{
			Name:        "get_exchange_rate",
			Description: "Get the exchange rate between two currencies",

			Parameters: jsonschema.Definition{
				Type: "object",

				Properties: map[string]jsonschema.Definition{
					"base_currency": {
						Type:        "string",
						Description: "The currency to convert from",
					},
					"target_currency": {
						Type:        "string",
						Description: "The currency to convert to",
					},
				},

				Required: []string{
					"base_currency",
					"target_currency",
				},
			},
		},
	}

	ctx := context.Background()

	prompt := "What is 1 CHF in USD?"
	println(prompt)

	resp1, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,

		Tools: []openai.Tool{
			toolGetExchangeRate,
		},

		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	if len(resp1.Choices[0].Message.ToolCalls) == 0 {
		panic("no tool calls")
	}

	fnName := resp1.Choices[0].Message.ToolCalls[0].Function.Name
	fnArgs := resp1.Choices[0].Message.ToolCalls[0].Function.Arguments

	println(fnName, fnArgs)

	if fnName != "get_exchange_rate" {
		panic("unexpected function name")
	}

	resp2, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,

		Tools: []openai.Tool{
			toolGetExchangeRate,
		},

		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleTool,

				ToolCallID: fnName,
				Content:    `{"symbol":"TSLA","company_name":"Tesla, Inc.","sector":"Consumer Cyclical","industry":"Auto Manufacturers","market_cap":611384164352,"pe_ratio":49.604652,"pb_ratio":9.762013,"dividend_yield":null,"eps":4.3,"beta":2.427,"52_week_high":299.29,"52_week_low":152.37}`,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	println(resp2.Choices[0].Message.Content)
}
