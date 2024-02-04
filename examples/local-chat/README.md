# Local Chat

## Run Example

- [Ollama](https://ollama.ai)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

Start Ollama Server

```shell
$ ollama start
```

Download [Mistral](https://mistral.ai) Model

```shell
$ ollama pull mistral
```

Start Example Application

```shell
docker compose up
```

## Open Web UI

```shell
$ open http://localhost:3000
```

## Completion API

The Completion API provides compatibility for the OpenAI API standard, allowing easier integrations into existing applications. (Documentation: https://platform.openai.com/docs/api-reference/chat/create)

```shell
curl http://localhost:8080/oai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "mistral",
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