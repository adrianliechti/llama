# Local Chat

## Run Example
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

Start Example Application

```shell
docker compose up --force-recreate --remove-orphans
```

## Open Web UI

```shell
$ open http://localhost:8000
```

## Completion API

The Completion API provides compatibility for the OpenAI API standard, allowing easier integrations into existing applications. (Documentation: https://platform.openai.com/docs/api-reference/chat/create)

```shell
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama",
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

## Embedding API

```shell
curl http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "nomic",
    "input": "Your text string goes here"
  }'
```