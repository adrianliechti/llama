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

| Category  | Document Types                                                                     |
|-----------|------------------------------------------------------------------------------------|
| Text      | `.txt`, `.eml`, `.msg`, `.html`, `.md`, `.rst`, `.rtf`                             |
| Images    | `.jpeg`, `.png`                                                                    |
| Documents | `.doc`, `.docx`, `.ppt`, `.pptx`, `.pdf`, `.odt`, `.epub`, `.csv`, `.tsv`, `.xlsx` |

```shell
curl http://localhost:8080/api/index/docs/unstructured \
  --header 'Content-Disposition: attachment; filename="presentation.pdf"' \
  --data-binary "@presentation.pdf"
```

```shell
go run . -url http://localhost:8080 -index docs -path $HOME/Documents
```

## Open Web UI

```shell
$ open http://localhost:3000
```