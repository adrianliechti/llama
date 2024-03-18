package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func main() {
	ctx := context.Background()

	config := openai.DefaultConfig("")
	config.BaseURL = "http://localhost:8080/oai/v1"

	client := openai.NewClientWithConfig(config)

	for {
		id := uuid.New().String()

		println("Press enter to start recording")

		waitForEnter()
		println("Recording...")

		timestamp := time.Now()
		record, recordCancel := context.WithCancel(ctx)

		go func() {
			for {
				time.Sleep(100 * time.Millisecond)

				if time.Since(timestamp) > 500*time.Millisecond {
					break
				}

			}

			println("Stop Recording")
			recordCancel()
		}()

		go func() {
			output := id + ".wav"

			args := []string{
				"-f", "avfoundation",
				"-i", ":0",
				"-ar", "16000",
				"-ac", "1",
				"-c:a", "pcm_s16le",
				"-flush_packets", "1",
				"-y",
				output,
			}

			defer os.Remove(output)

			ffmpeg := exec.CommandContext(record, "ffmpeg", args...)
			//ffmpeg.Stdout = os.Stdout
			//ffmpeg.Stderr = os.Stderr
			ffmpeg.Run()

			println("Recording done")
			transcribe(ctx, client, output)
		}()

		for {
			if record.Err() != nil {
				break
			}

			waitForEnter()
			timestamp = time.Now()
		}
	}

}

func waitForEnter() {
	var b []byte = make([]byte, 1)

	for {
		os.Stdin.Read(b)

		if len(b) == 0 || b[0] == 10 {
			break
		}
	}
}

func transcribe(ctx context.Context, client *openai.Client, file string) (string, error) {
	transcription, err := client.CreateTranscription(ctx, openai.AudioRequest{
		Model:    "whisper",
		FilePath: file,
	})

	if err != nil {
		return "", err
	}

	println("> " + transcription.Text)

	complete(ctx, client, transcription.Text)

	return "", nil
}

func complete(ctx context.Context, client *openai.Client, prompt string) (string, error) {
	stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:     "mistral",
		MaxTokens: 20,

		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})

	if err != nil {
		return "", err
	}

	var result string

	print("< ")

	for {
		response, err := stream.Recv()

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			break
		}

		text := response.Choices[0].Delta.Content
		result += text

		print(text)
	}

	return strings.TrimSpace(result), nil
}
