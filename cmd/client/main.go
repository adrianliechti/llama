package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/adrianliechti/llama/config"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func main() {
	urlFlag := flag.String("url", "http://localhost:8080/v1", "server url")
	tokenFlag := flag.String("token", "", "server token")
	modelFlag := flag.String("model", "", "model id")

	flag.Parse()

	ctx := context.Background()

	cfg := openai.DefaultConfig(*tokenFlag)
	cfg.BaseURL = *urlFlag

	client := openai.NewClientWithConfig(cfg)
	model := *modelFlag

	if model == "" {
		val, err := selectModel(ctx, client)

		if err != nil {
			panic(err)
		}

		model = val
	}

	if config.DetectModelType(model) == config.ModelTypeEmbedder {
		embed(ctx, client, model)
		return
	}

	if config.DetectModelType(model) == config.ModelTypeRenderer {
		render(ctx, client, model)
		return
	}

	if config.DetectModelType(model) == config.ModelTypeSynthesizer {
		synthesize(ctx, client, model)
		return
	}

	chat(ctx, client, model)
}

func selectModel(ctx context.Context, client *openai.Client) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	output := os.Stdout

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

	output.WriteString("\n")

	model := list.Models[idx-1].ID
	return model, nil
}

func chat(ctx context.Context, client *openai.Client, model string) {
	reader := bufio.NewReader(os.Stdin)
	output := os.Stdout

	var messages []openai.ChatCompletionMessage

LOOP:
	for {
		output.WriteString(">>> ")
		input, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		input = strings.TrimSpace(input)

		if strings.HasPrefix(input, "/") {
			switch strings.ToLower(input) {
			case "/reset":
				messages = nil
				continue LOOP

			case "/repeat":
				if len(messages) == 0 {
					continue LOOP
				}

				input = messages[len(messages)-1].Content
				messages = messages[:len(messages)-1]

			default:
				output.WriteString("Unknown command\n")
				continue LOOP
			}
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: input,
		})

		req := openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		}

		stream, err := client.CreateChatCompletionStream(ctx, req)

		if err != nil {
			output.WriteString(err.Error() + "\n")
			continue LOOP
		}

		defer stream.Close()

		var buffer strings.Builder

		for {
			resp, err := stream.Recv()

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				output.WriteString(err.Error() + "\n")
				continue LOOP
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

func embed(ctx context.Context, client *openai.Client, model string) {
	reader := bufio.NewReader(os.Stdin)
	output := os.Stdout

LOOP:
	for {
		output.WriteString(">>> ")
		input, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		input = strings.TrimSpace(input)

		req := openai.EmbeddingRequest{
			Input: input,
			Model: openai.EmbeddingModel(model),
		}

		resp, err := client.CreateEmbeddings(ctx, req)

		if err != nil {
			output.WriteString(err.Error() + "\n")
			continue LOOP
		}

		embeddings := resp.Data[0].Embedding

		for i, e := range embeddings {
			if i > 0 {
				output.WriteString(", ")
			}
			output.WriteString(fmt.Sprintf("%f", e))
		}

		output.WriteString("\n")
		output.WriteString("\n")
	}
}

func render(ctx context.Context, client *openai.Client, model string) {
	reader := bufio.NewReader(os.Stdin)
	output := os.Stdout

LOOP:
	for {
		output.WriteString(">>> ")
		input, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		input = strings.TrimSpace(input)

		req := openai.ImageRequest{
			Prompt: input,
			Model:  model,
			//Size:           openai.CreateImageSize1024x1024,
			ResponseFormat: openai.CreateImageResponseFormatB64JSON,
			N:              1,
		}

		resp, err := client.CreateImage(ctx, req)

		if err != nil {
			output.WriteString(err.Error() + "\n")
			continue LOOP
		}

		data, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)

		if err != nil {
			output.WriteString(err.Error() + "\n")
			continue LOOP
		}

		name := uuid.New().String()

		if ext, _ := mime.ExtensionsByType(http.DetectContentType(data)); len(ext) > 0 {
			name += ext[0]
		}

		os.WriteFile(name, data, 0600)
		fmt.Println("Saved: " + name)

		output.WriteString("\n")
		output.WriteString("\n")
	}
}

func synthesize(ctx context.Context, client *openai.Client, model string) {
	reader := bufio.NewReader(os.Stdin)
	output := os.Stdout

LOOP:
	for {
		output.WriteString(">>> ")
		input, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		input = strings.TrimSpace(input)

		req := openai.CreateSpeechRequest{
			Input: input,
			Model: openai.SpeechModel(model),

			ResponseFormat: openai.SpeechResponseFormatWav,
		}

		resp, err := client.CreateSpeech(ctx, req)

		if err != nil {
			output.WriteString(err.Error() + "\n")
			continue LOOP
		}

		data, err := io.ReadAll(resp)

		resp.Close()

		if err != nil {
			output.WriteString(err.Error() + "\n")
			continue LOOP
		}

		name := uuid.New().String()

		if ext, _ := mime.ExtensionsByType(http.DetectContentType(data)); len(ext) > 0 {
			name += ext[0]
		}

		os.WriteFile(name, data, 0600)
		fmt.Println("Saved: " + name)

		output.WriteString("\n")
		output.WriteString("\n")
	}
}
