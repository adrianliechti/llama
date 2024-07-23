# Local Chat using LangChain

## Run Example

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

Start Example Application

```shell
docker compose up
```

## Open Web UI

```shell
$ open http://localhost:8501
```

## Completion API

The Completion API provides compatibility for the OpenAI API standard, allowing easier integrations into existing applications. (Documentation: https://platform.openai.com/docs/api-reference/chat/create)

```shell
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "custom",
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