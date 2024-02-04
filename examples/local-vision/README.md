# Local Vision

## Run Example

- [Ollama](https://ollama.ai)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

Start Ollama Server

```shell
$ ollama start
```

Download [LLaVA](https://llava-vl.github.io) Model

```shell
$ ollama pull llava
```

Start Example Application

```shell
docker compose up
```

## Open Web UI

```shell
$ open http://localhost:3000
```

## Vision API

The Vision API provides compatibility for the OpenAI API standard, allowing easier integrations into existing applications. (Documentation: https://platform.openai.com/docs/guides/vision)

```shell
curl http://localhost:8080/oai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llava",
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