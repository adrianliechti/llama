# OpenAI Adapter

- Text Generation (ChatGPT)
- Image Recognition (Vision)
- Audio Transcriptions (Whisper)

```bash
export OPENAI_API_KEY=sk-......

docker compose up --force-recreate --remove-orphans
```

open [localhost:8501](http://localhost:8501) in your favorite browser

## Completion API

https://platform.openai.com/docs/api-reference/chat/create

```shell
curl http://localhost:8080/oai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-turbo",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Hello!"
      }
    ]
  }'
```

## Vision API

https://platform.openai.com/docs/guides/vision

```shell
curl http://localhost:8080/oai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4-turbo",
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "text",
            "text": "What’s in this image?"
          },
          {
            "type": "image_url",
            "image_url": {
              "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
            }
          }
        ]
      }
    ]
  }'
```

## Transcription API

```shell
# Download Sample File
curl -o jfk.wav https://github.com/ggerganov/whisper.cpp/raw/master/samples/jfk.wav

# Run Transcriptions
curl http://localhost:8080/oai/v1/audio/transcriptions \
  -H "Content-Type: multipart/form-data" \
  -F file="@jfk.wav" \
  -F model="whisper-1"
```