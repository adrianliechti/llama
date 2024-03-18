# Local Voice Bot

## Run Example

- [Whisper.cpp](https://github.com/ggerganov/whisper.cpp)
- [Llama.cpp](https://github.com/ggerganov/llama.cpp)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)


Start LLama & Whisper Server

```shell
task llama:server
task whisper:server
```

Start Example Application

```shell
docker compose up
```

## Transcription API

The Transcription API provides compatibility for the OpenAI API standard, allowing easier integrations into existing applications. (Documentation: https://platform.openai.com/docs/api-reference/audio/createTranscription)

```shell
# Download Sample File
curl -o jfk.wav https://github.com/ggerganov/whisper.cpp/raw/master/samples/jfk.wav

# Run Transcriptions
curl http://localhost:8080/oai/v1/audio/transcriptions \
  -H "Content-Type: multipart/form-data" \
  -F file="@jfk.wav" \
  -F model="whisper"
```