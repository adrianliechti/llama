# Local Voice Bot

## Run Example

- [Llama.cpp](https://github.com/ggerganov/llama.cpp)
- [Whisper.cpp](https://github.com/ggerganov/whisper.cpp)
- [Mimic](https://github.com/MycroftAI/mimic3)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)


Start LLama & Whisper Server

```shell
task llama:server
task whisper:server
```

Start Mimic Server

```
mkdir -p mimic3
chmod 777 mimic3
docker run -it -p 59125:59125 -v $(pwd)/mimic3:/home/mimic3/.local/share/mycroft/mimic3 mycroftai/mimic3
```

Start Example Application

```shell
docker compose up
```

## Transcription API

The Transcription API provides compatibility for the OpenAI API standard, allowing easier integrations into existing applications. (Documentation: https://platform.openai.com/docs/api-reference/audio/createTranscription)

```shell
# Download Sample File
curl -Lo jfk.wav https://github.com/ggerganov/whisper.cpp/raw/master/samples/jfk.wav

# Run Transcriptions
curl http://localhost:8080/v1/audio/transcriptions \
  -H "Content-Type: multipart/form-data" \
  -F file="@jfk.wav" \
  -F model="whisper-1"
```