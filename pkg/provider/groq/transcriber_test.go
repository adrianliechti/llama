package groq_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/adrianliechti/llama/pkg/provider"
	"github.com/adrianliechti/llama/pkg/provider/groq"

	"github.com/stretchr/testify/require"
)

func TestTranscriber(t *testing.T) {
	ctx := context.Background()
	token := os.Getenv("GROQ_API_TOKEN")
	model := "whisper-large-v3"

	if token == "" {
		t.Skip("GROQ_API_TOKEN required for this test")
	}

	p, err := groq.NewTranscriber("", model, groq.WithToken(token))
	require.NoError(t, err)

	resp, err := http.Get("https://github.com/ggerganov/whisper.cpp/raw/master/samples/jfk.wav")
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	result, err := p.Transcribe(ctx, provider.File{
		Name: "jfk.wav",

		Content:     bytes.NewReader(data),
		ContentType: "audio/wav",
	}, nil)

	require.NoError(t, err)
	require.NotEmpty(t, result.Content)

	t.Log(result.Content)
}
