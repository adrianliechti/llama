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
$ ollama pull llama3.1
$ ollama pull nomic-embed-text
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
curl http://localhost:8080/v1/index/docs/unstructured \
  --header 'Content-Disposition: attachment; filename="presentation.pdf"' \
  --data-binary "@presentation.pdf"
```

```shell
docker run -it --rm -v ./:/docs -w /docs ghcr.io/adrianliechti/llama-platform /ingest -url http://host.docker.internal:8080
```

## Open Web UI

```shell
$ open http://localhost:8501
```