# Run Platform in Docker / CUDA

```shell
curl http://localhost:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "input": "Hello!",
    "model": "nomic-embed-text"
  }'
```

```shell
curl http://localhost:8080/v1/chat/completions \
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