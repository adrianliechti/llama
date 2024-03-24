# Run Platform in Docker / CUDA

```shell
mkdir /models
chmod 777 /models

curl -Lo /models/whisper-ggml-medium.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium.bin
```

```shell
curl http://localhost:8080/oai/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello!",
    "model": "nomic-embed-text"
  }'
```

```shell
curl http://localhost:8080/oai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "mistral-7b-instruct",
    "messages": [
      {
        "role": "user",
        "content": "Hello!"
      }
    ]
  }'
```

```shell
curl http://localhost:8080/oai/v1/audio/speech \
  -H "Content-Type: application/json" \
  -d '{
    "model": "tts-1",
    "input": "The quick brown fox jumped over the lazy dog.",
    "voice": "en"
  }' \
  --output speech.wav
```

```shell
curl http://localhost:8080/oai/v1/audio/transcriptions \
  -H "Content-Type: multipart/form-data" \
  -F file="@speech.wav" \
  -F model="whisper-1"
```