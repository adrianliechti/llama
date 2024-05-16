# Anthropic Chat

```shell
export HF_TOKEN=hf_......

docker compose up --force-recreate --remove-orphans
```

open [localhost:8501](http://localhost:8501) in your favorite browser


```shell
curl http://localhost:8080/oai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3-8b-instruct",
    "messages": [
      {
        "role": "user",
        "content": "Hello!"
      }
    ]
  }'
```

```shell
curl http://localhost:8080/oai/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "all-mpnet-base-v2",
    "input": "Hello!"
  }'
```