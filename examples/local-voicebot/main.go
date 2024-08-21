package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func main() {
	ctx := context.Background()

	config := openai.DefaultConfig("")
	config.BaseURL = "http://localhost:8080/v1"

	completionModel := "llama"
	synthesizerModel := "tts-1"
	transcriptionModel := "whisper-1"

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

			transcription, err := transcribe(ctx, client, transcriptionModel, output)

			if err != nil {
				println(err.Error())
				return
			}

			prompt := transcription.Text
			println("> " + prompt)

			if prompt == "" || prompt == "[BLANK_AUDIO]" {
				return
			}

			if transcription.Language != "" {
				println("Language: " + transcription.Language)
				prompt = strings.TrimRight(prompt, ".") + ". Answer in " + transcription.Language + "."
			}

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			})

			completion, err := complete(ctx, client, completionModel, messages)

			if err != nil {
				println(err.Error())
				return
			}

			answer := completion.Choices[0].Message.Content

			answer = regexp.MustCompile(`\[.*?\]|\(.*?\)`).ReplaceAllString(answer, "")
			answer = regexp.MustCompile(`\[.*?\]?|\(.*?\)?`).ReplaceAllString(answer, "")
			answer = strings.Split(answer, "\n")[0]

			println("< " + answer)

			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: answer,
			})

			data, err := synthesize(ctx, client, synthesizerModel, transcription.Language, answer)

			if err != nil {
				println(err.Error())
				return
			}

			f, err := os.CreateTemp("", "voicebot-*.wav")

			if err != nil {
				println(err.Error())
				return
			}

			if _, err := io.Copy(f, data); err != nil {
				println(err.Error())
				return
			}

			f.Close()

			play(ctx, f.Name())

			//say(ctx, answer)
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

func transcribe(ctx context.Context, client *openai.Client, model, file string) (openai.AudioResponse, error) {
	return client.CreateTranscription(ctx, openai.AudioRequest{
		Model:    model,
		FilePath: file,
	})
}

func complete(ctx context.Context, client *openai.Client, model string, messages []openai.ChatCompletionMessage) (openai.ChatCompletionResponse, error) {
	return client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	})
}

func synthesize(ctx context.Context, client *openai.Client, model, voice, text string) (io.ReadCloser, error) {
	return client.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Input: text,

		Model: openai.SpeechModel(model),
		Voice: openai.SpeechVoice(voice),
	})
}

func play(ctx context.Context, path string) error {
	if runtime.GOOS == "darwin" {
		return exec.CommandContext(ctx, "afplay", path).Run()
	}

	if runtime.GOOS == "linux" {
		return exec.CommandContext(ctx, "aplay", path).Run()
	}

	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "powershell", "-c", "(New-Object Media.SoundPlayer \""+path+"\").PlaySync()").Run()
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
