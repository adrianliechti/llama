
# LLAMA Gateway

The LLAMA Gateway provides an unified API to various Large Language Models (LLM) and higher level functionality like 
Retrieval-Augmented Generation (RAG) for use cases like:

- Enterprise Chat
- Question/Answering (QA) over Documents and Code


## Integrations

###  Large Language Models (LLM)

- OpenAI API (or compatible)  
  - [OpenAI Platform](https://platform.openai.com/docs/introduction)
  - [Azure OpenAI Service](https://azure.microsoft.com/en-us/products/ai-services/openai-service)
  - [vLLM](https://docs.vllm.ai)
  - ...

- Local Models
  - [LLAMA.CPP](https://github.com/ggerganov/llama.cpp) Server

- Embedding Models
  -  [Sentence BERT](https://www.sbert.net) 


### Vector Indexes

- [Chroma](https://www.trychroma.com) Embedding Database
- [Weaviate](https://weaviate.io) Vector Database
- In-Memory cosine similarity

### User Authorizers

- Static Token
- OIDC JWT Tokens


## Example Application

The Docker `compose.yaml` file starts a simple web-based [Chatbot UI](https://github.com/mckaywrigley/chatbot-ui)  (port http/3000) and a [LLAMA.CPP](https://github.com/ggerganov/llama.cpp) server. The LLAMA Gateway API is exposed locally (port http/8080) using the static token `changeme`. An OpenAI-compatible API is availabe at http://localhost:8080/oai/v1. 

While starting up, a [Mistral 7B Instruct](https://mistral.ai/news/announcing-mistral-7b/) model file will be downloaded from [Hugging Face](https://huggingface.co) (see ./models) if not already exists.

The sample also starts a sentence-transformer server to provide embedding APIs.

For broad compatibility with existing tools (like the bundled WebUI), the models are aliased as `gpt-3.5-turbo` and `text-embedding-ada-002`.

Run example application

```bash
$ docker compose up
```

Browse to http://localhost:3000

## Configuration

### Providers

#### OpenAI Platform

```yaml
providers:
  - type: openai
    token: sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    models:
      gpt-3.5-turbo:
        id: gpt-3.5-turbo-1106

      gpt-4:
        id: gpt-4-1106-preview
        
      text-embedding-ada-002:
        id: text-embedding-ada-002
```

#### Azure OpenAI Service

```yaml
providers:
  - type: openai
    url: https://xxxxxxxx.openai.azure.com
    token: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

    models:
      gpt-3.5-turbo:
        id: gpt-35-turbo-16k

      gpt-4:
        id: gpt-4-32k
        
      text-embedding-ada-002:
        id: text-embedding-ada-002
```

#### LLAMA.CPP

```shell
server --port 9081 --model ./models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
```

```shell
docker run -it --rm -p 9081:9081 -v ./models/:/models/ ghcr.io/ggerganov/llama.cpp:full --server --host 0.0.0.0 --port 9081 --model /models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
```

```yaml
providers:
  - type: llama
    url: http://localhost:9081

    models:
      mistral-7b-instruct:
        id: /models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
        template: mistral
```

#### Sentence-BERT

```shell
docker run -it --rm -p 9082:8080 semitechnologies/transformers-inference:sentence-transformers-multi-qa-MiniLM-L6-cos-v1
```

```yaml
providers:
  - type: sbert
    url: http://localhost:9082

    models:
      multi-qa-minilm-l6-cos-v1:
        id: multi-qa-minilm-l6-cos-v1
```

### Indexes

#### Chroma

```shell
docker run -it --rm -p 9083:8000 ghcr.io/chroma-core/chroma
```

```yaml
indexes:
  docs:
    type: chroma
    url: http://localhost:9083
    namespace: docs
    embedding: text-embedding-ada-002
```

#### Weaviate

```shell
docker run -it --rm -p 9084:8080 -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true -e PERSISTENCE_DATA_PATH=/data semitechnologies/weaviate
```

```yaml
indexes:
  docs:
    type: weaviate
    url: http://localhost:9084
    namespace: Document
    embedding: multi-qa-minilm-l6-cos-v1  
```

#### In-Memory

```yaml
indexes:
  docs:
    type: memory   
    embedding: text-embedding-ada-002
```

## Use Cases

### Retrieval Augmented Generation (RAG)

#### Index Documents

```
POST http://localhost:8080/api/index/{index-name}
```

```json
[
    {
        "id": "id1",
        "content": "content of document..."
    },
    {
        "id": "id2",
        "content": "content of document..."
    }
]
```