# Local Audio Transcription

## Run Example

- [Whisper.cpp](https://github.com/ggerganov/whisper.cpp)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

Download Model

```shell
$ curl -Lo whisper-ggml-medium.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-medium.bin
```

Start Whisper Server

```shell
$ ./server --port 8000 --model whisper-ggml-medium.bin
```

Start Example Application

```shell
docker compose up
```

## Transcribe File

```shell
# Download Sample File
curl -o jfk.wav https://github.com/ggerganov/whisper.cpp/raw/master/samples/jfk.wav

# Run Transcriptions
curl http://localhost:8080/oai/v1/audio/transcriptions \
  -H "Content-Type: multipart/form-data" \
  -F file="@jfk.wav" \
  -F model="whisper"
```