package main

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func main() {
	ctx := context.Background()

	config := openai.DefaultConfig("")
	config.BaseURL = "http://localhost:8080/oai/v1"

	completionModel := "mistral"
	transcriptionModel := "whisper"

	client := openai.NewClientWithConfig(config)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a voice assistant. Answer short and concise!",
		},
	}

	println("Press enter to start / stop recording")

	for {
		waitForEnter()
		println("Recording...")

		recordCtx, recordCancel := context.WithCancel(ctx)

		go func() {
			output, err := record(recordCtx)

			if err != nil {
				println(err.Error())
				return
			}

			defer os.Remove(output)

			transcription, _ := transcribe(ctx, client, transcriptionModel, output)
			println("> " + transcription)

			if transcription == "" || transcription == "[BLANK_AUDIO]" {
				return
			}

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: transcription,
			})

			completion, _ := complete(ctx, client, completionModel, messages)
			println("< " + completion)

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: completion,
			})

			say(ctx, completion)
		}()

		waitForEnter()
		recordCancel()
	}
}

func record(ctx context.Context) (string, error) {
	output := strings.ReplaceAll(uuid.NewString(), "-", "") + ".wav"

	args := []string{
		"-loglevel", "-0",
		"-f", "avfoundation",
		"-i", ":0",
		"-ar", "16000",
		"-ac", "1",
		"-c:a", "pcm_s16le",
		"-flush_packets", "1",
		"-y",
		output,
	}

	ffmpeg := exec.Command("ffmpeg", args...)
	//ffmpeg.Stdout = os.Stdout
	//ffmpeg.Stderr = os.Stderr

	go func() {
		<-ctx.Done()
		ffmpeg.Process.Signal(os.Interrupt)
	}()

	ffmpeg.Run()

	return output, nil
}

func transcribe(ctx context.Context, client *openai.Client, model, file string) (string, error) {
	transcription, err := client.CreateTranscription(ctx, openai.AudioRequest{
		Model:    model,
		FilePath: file,
	})

	if err != nil {
		return "", err
	}

	return transcription.Text, nil
}

func complete(ctx context.Context, client *openai.Client, model string, messages []openai.ChatCompletionMessage) (string, error) {
	completion, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,

		Messages: messages,
	})

	if err != nil {
		return "", err
	}

	return completion.Choices[0].Message.Content, nil
}

func say(ctx context.Context, text string) error {
	if runtime.GOOS == "darwin" {
		exec.CommandContext(ctx, "say", text).Run()
	}

	return nil
}

func waitForEnter() {
	var b []byte = make([]byte, 1)

	for {
		os.Stdin.Read(b)

		if len(b) == 0 {
			continue
		}

		if b[0] == 10 {
			break
		}
	}
}
