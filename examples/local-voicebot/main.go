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
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	chatmodel := "gpt-4o"
	audiomodel := "whisper-1"
	speakmodel := "tts-1"

	url := os.Getenv("OPENAI_API_BASE")

	if url == "" {
		url = "http://localhost:8080/v1"
	}

	url = strings.TrimRight(url, "/") + "/"

	options := []option.RequestOption{
		option.WithBaseURL(url),
	}

	client := openai.NewClient(options...)

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("Your knowledge cutoff is 2023-10. You are a helpful, witty, and friendly AI. Act like a human, but remember that you aren't a human and that you can't do human things in the real world. Your voice and personality should be warm and engaging, with a lively and playful tone. If interacting in a non-English language, start by using the standard accent or dialect familiar to the user. Talk quickly. You should always call a function if you can. Answer as briefly and concisely as possible. Keep it short."),
	}

	for ctx.Err() == nil {
		println("🙉 Listening...")

		data, err := recordChunk(ctx)

		if err != nil {
			println("error:", err.Error())
			continue
		}

		transcription, err := client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
			Model: openai.F(audiomodel),
			File:  openai.FileParam(bytes.NewReader(data), "file.wav", "audio/wav"),
		})

		if err != nil {
			println("error:", err.Error())
			continue
		}

		fmt.Println("💬 " + transcription.Text)

		if strings.TrimSpace(transcription.Text) == "" {
			continue
		}

		messages = append(messages, openai.UserMessage(transcription.Text))

		stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    openai.F(chatmodel),
			Messages: openai.F(messages),
		})

		print("📣 ")

		var text string

		for stream.Next() {
			chunk := stream.Current()

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				text += content

				print(content)
			}
		}

		println()

		if err := stream.Err(); err != nil {
			println("error:", err.Error())
			continue
		}

		messages = append(messages, openai.AssistantMessage(text))

		sayText(ctx, client, speakmodel, text)
	}
}

func sayText(ctx context.Context, client *openai.Client, model, input string) error {
	result, err := client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Model: openai.F(model),
		Input: openai.F(input),

		Voice:          openai.F(openai.AudioSpeechNewParamsVoiceAlloy),
		ResponseFormat: openai.F(openai.AudioSpeechNewParamsResponseFormatWAV),
	})

	if err != nil {
		return err
	}

	defer result.Body.Close()

	path := filepath.Join(os.TempDir(), uuid.New().String()+".mp3")
	defer os.Remove(path)

	file, err := os.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()
	defer os.Remove(path)

	if _, err := io.Copy(file, result.Body); err != nil {
		return err
	}

	file.Close()

	if err := playFile(ctx, path); err != nil {
		return err
	}

	return nil
}

func playFile(ctx context.Context, path string) error {
	if _, err := exec.LookPath("play"); err == nil {
		return playFileSOX(ctx, path)
	}

	if _, err := exec.LookPath("ffplay"); err == nil {
		return playFileFFMPEG(ctx, path)
	}

	return errors.New("neither FFmpeg nor SoX are installed")
}

func playFileSOX(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "play", path)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func playFileFFMPEG(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "ffplay", "-autoexit", "-nodisp", path)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func recordChunk(ctx context.Context) ([]byte, error) {
	if _, err := exec.LookPath("sox"); err == nil {
		return recordChunkSOX(ctx)
	}

	if _, err := exec.LookPath("ffmpeg"); err == nil {
		return recordChunkFFMPEG(ctx)
	}

	return nil, errors.New("neither FFmpeg nor SoX are installed")
}

func recordChunkSOX(ctx context.Context) ([]byte, error) {
	path := filepath.Join(os.TempDir(), uuid.New().String()+".wav")
	defer os.Remove(path)

	args := []string{
		"-d",
		path,
		"silence",
		"1", "0.0", "1%",
		"1", "1.5", "1%",
	}

	cmd := exec.CommandContext(ctx, "sox", args...)

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("error reading file:", err)
		return nil, err
	}

	return data, nil
}

func recordChunkFFMPEG(ctx context.Context) ([]byte, error) {
	var args []string

	path := filepath.Join(os.TempDir(), uuid.New().String()+".wav")
	defer os.Remove(path)

	switch runtime.GOOS {
	case "darwin":
		args = []string{
			"-f", "avfoundation",
			"-i", ":0",
			"-af", "silencedetect=noise=-30dB:d=1",
			path,
		}
	case "windows":
		args = []string{
			"-f", "dshow",
			"-i", "audio=default",
			"-af", "silencedetect=noise=-30dB:d=1",
			path,
		}
	case "linux":
		args = []string{
			"-f", "alsa",
			"-i", "default",
			"-af", "silencedetect=noise=-30dB:d=1",
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

	err = cmd.Process.Signal(os.Interrupt)

	if err != nil {
		err = cmd.Process.Kill()
	}

	if err != nil {
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
