# Local RAG

## Run Example
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

Start Example Application

```shell
docker compose up --force-recreate --remove-orphans
```

## Index Documents

```shell
docker run -it --rm -v ./:/data -w /data --pull=always ghcr.io/adrianliechti/llama-platform /ingest -url http://host.docker.internal:8080 -token -
```

## Verify Documents

```
open http://localhost:8080/v1/index/docs
open http://localhost:6333/dashboard#/collections/docs
```

## Open Web UI

```shell
$ open http://localhost:8000
```
