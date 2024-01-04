package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func main() {
	llamaURL := os.Getenv("LLAMA_URL")
	llamaToken := os.Getenv("LLAMA_TOKEN")

	model := os.Getenv("LLAMA_MODEL")

	if llamaURL == "" {
		llamaURL = "http://localhost:8080/oai/v1"
	}

	ctx := context.Background()

	reader := bufio.NewReader(os.Stdin)
	output := os.Stdout

	config := openai.DefaultConfig(llamaToken)
	config.BaseURL = llamaURL

	client := openai.NewClientWithConfig(config)

	if model == "" {
		list, err := client.ListModels(ctx)

		if err != nil {
			panic(err)
		}

		sort.SliceStable(list.Models, func(i, j int) bool {
			return list.Models[i].ID < list.Models[j].ID
		})

		for i, m := range list.Models {
			output.WriteString(fmt.Sprintf("%2d) ", i+1))
			output.WriteString(m.ID)
			output.WriteString("\n")
		}

		output.WriteString(" >  ")
		sel, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		idx, err := strconv.Atoi(strings.TrimSpace(sel))

		if err != nil {
			panic(err)
		}

		model = list.Models[idx-1].ID
		output.WriteString("\n")
	}

	var messages []openai.ChatCompletionMessage

	for {
		output.WriteString(">>> ")
		input, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: strings.TrimSpace(input),
		})

		req := openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		}

		stream, err := client.CreateChatCompletionStream(ctx, req)

		if err != nil {
			panic(err)
		}

		defer stream.Close()

		var buffer strings.Builder

		for {
			resp, err := stream.Recv()

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				panic(err)
			}

			content := resp.Choices[0].Delta.Content

			buffer.WriteString(content)
			output.WriteString(content)
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: strings.TrimSpace(buffer.String()),
		})

		output.WriteString("\n")
		output.WriteString("\n")
	}
}
