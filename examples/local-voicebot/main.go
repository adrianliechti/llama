package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	chatmodel := "gpt-4o"
	audiomodel := "whisper-1"
	speakmodel := "tts-1-hd"

	url := os.Getenv("OPENAI_API_BASE")
	token := os.Getenv("OPENAI_API_KEY")

	config := openai.DefaultConfig(token)
	config.BaseURL = "http://localhost:8080/v1"

	if url != "" {
		config.BaseURL = url
	}

	client := openai.NewClientWithConfig(config)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Your knowledge cutoff is 2023-10. You are a helpful, witty, and friendly AI. Act like a human, but remember that you aren't a human and that you can't do human things in the real world. Your voice and personality should be warm and engaging, with a lively and playful tone. If interacting in a non-English language, start by using the standard accent or dialect familiar to the user. Talk quickly. You should always call a function if you can. Answer as briefly and concisely as possible.",
		},
	}

	for ctx.Err() == nil {
		println("🙉 Listening...")

		data, err := recordChunk(ctx)

		if err != nil {
			println("error:", err.Error())
			continue
		}

		response, err := client.CreateTranscription(ctx, openai.AudioRequest{
			Model: audiomodel,

			FilePath: "chunk.wav",
			Reader:   bytes.NewReader(data),
		})

		if err != nil {
			println("error:", err.Error())
			continue
		}

		fmt.Println("💬 " + response.Text)

		if strings.TrimSpace(response.Text) == "" {
			continue
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: response.Text,
		})

		result, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    chatmodel,
			Messages: messages,
		})

		if err != nil {
			println("error:", err.Error())
			continue
		}

		message := result.Choices[0].Message
		messages = append(messages, message)

		println("📣 " + message.Content)

		sayText(ctx, client, speakmodel, message.Content)
	}
}

func sayText(ctx context.Context, client *openai.Client, model, text string) error {
	data, err := client.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model: openai.SpeechModel(model),
		Input: text,
	})

	if err != nil {
		return err
	}

	path := filepath.Join(os.TempDir(), uuid.New().String()+".mp3")
	defer os.Remove(path)

	file, err := os.Create(path)

	if err != nil {
		return err
	}

	io.Copy(file, data)

	defer file.Close()
	defer os.Remove(path)

	cmd := exec.CommandContext(ctx, "ffplay", "-autoexit", "-nodisp", path)
	cmd.Run()

	return nil
}

func recordChunk(ctx context.Context) ([]byte, error) {
	var args []string

	path := filepath.Join(os.TempDir(), uuid.New().String()+".wav")
	defer os.Remove(path)

	switch runtime.GOOS {
	case "darwin":
		args = []string{
			"-f", "avfoundation",
			"-i", ":0",
			"-af", "silencedetect=noise=-30dB:d=2",
			path,
		}
	case "windows":
		args = []string{
			"-f", "dshow",
			"-i", "audio=default",
			"-af", "silencedetect=noise=-30dB:d=2",
			path,
		}
	case "linux":
		args = []string{
			"-f", "alsa",
			"-i", "default",
			"-af", "silencedetect=noise=-30dB:d=2",
			path,
		}
	}

	if len(args) == 0 {
		return nil, errors.New("unsupported platform")
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024)
	silencePattern := regexp.MustCompile(`silence_start`)

	for {
		n, err := stderr.Read(buffer)

		if err != nil {
			break
		}

		output := string(buffer[:n])

		if silencePattern.MatchString(output) {
			break
		}
	}

	if err := cmd.Process.Kill(); err != nil {
		fmt.Println("Error killing FFmpeg process:", err)
		return nil, err
	}

	cmd.Process.Wait()

	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("error reading file:", err)
		return nil, err
	}

	return data, nil
}
