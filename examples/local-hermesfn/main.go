package main

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func main() {
	config := openai.DefaultConfig("")
	config.BaseURL = "http://localhost:8080/v1"

	model := "hermes-2-pro"

	client := openai.NewClientWithConfig(config)

	toolGetStockFundamentals := openai.Tool{
		Type: openai.ToolTypeFunction,

		Function: &openai.FunctionDefinition{
			Name: "get_stock_fundamentals",
			Description: `get_stock_fundamentals(symbol: str) -> dict - Get fundamental data for a given stock symbol using yfinance API.
    Args:
        symbol (str): The stock symbol.
    
    Returns:
        dict: A dictionary containing fundamental data.
            Keys:
                - 'symbol': The stock symbol.
                - 'company_name': The long name of the company.
                - 'sector': The sector to which the company belongs.
                - 'industry': The industry to which the company belongs.
                - 'market_cap': The market capitalization of the company.
                - 'pe_ratio': The forward price-to-earnings ratio.
                - 'pb_ratio': The price-to-book ratio.
                - 'dividend_yield': The dividend yield.
                - 'eps': The trailing earnings per share.
                - 'beta': The beta value of the stock.
                - '52_week_high': The 52-week high price of the stock.
                - '52_week_low: The 52-week low price of the stock.`,

			Parameters: jsonschema.Definition{
				Type: "object",

				Properties: map[string]jsonschema.Definition{
					"symbol": {
						Type: "string",
					},
				},

				Required: []string{
					"symbol",
				},
			},
		},
	}

	ctx := context.Background()

	prompt := "Fetch the stock fundamentals data for Tesla (TSLA)"
	println(prompt)

	resp1, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,

		Tools: []openai.Tool{
			toolGetStockFundamentals,
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

	if fnName != "get_stock_fundamentals" {
		panic("unexpected function name")
	}

	resp2, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,

		Tools: []openai.Tool{
			toolGetStockFundamentals,
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
