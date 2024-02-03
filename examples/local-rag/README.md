# Local RAG

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

## Upload Documents

```shell
curl http://localhost:8080/api/index/docs/unstructured \
  --header 'Content-Disposition: attachment; filename="presentation.pdf"' \
  --data-binary "@presentation.pdf"
```

## Open Web UI

```shell
$ open http://localhost:3000
```