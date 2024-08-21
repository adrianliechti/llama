# Local Voice Bot

## Run Example

- [Llama.cpp](https://github.com/ggerganov/llama.cpp)
- [Whisper.cpp](https://github.com/ggerganov/whisper.cpp)
- [Mimic](https://github.com/MycroftAI/mimic3)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)


### Start Providers

```shell
task llama:server
task whisper:server
task tts:server
```

### Start Example Application

```shell
docker compose up
```

## Transcription API

The Transcription API provides compatibility for the OpenAI API standard, allowing easier integrations into existing applications. (Documentation: https://platform.openai.com/docs/api-reference/audio/createTranscription)

```shell
# Download Sample File
curl -Lo jfk.wav https://github.com/ggerganov/whisper.cpp/raw/master/samples/jfk.wav

# Run Audio Transcriptions
curl http://localhost:8080/v1/audio/transcription \
  -H "Content-Type: multipart/form-data" \
  -F file="@jfk.wav" \
  -F model="whisper-1"

### Run Audio Speech
curl http://localhost:8080/v1/audio/speech \
  -H "Content-Type: application/json" \
  -d '{
    "model": "tts-1",
    "input": "Today is a wonderful day to build something people love!",
    "voice": "alloy"
  }' \
  --output speech.mp3
```