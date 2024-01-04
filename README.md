
# LLAMA Platform

The LLAMA Platform provides an unified API to various Large Language Models (LLM) and higher level functionality like 
Retrieval-Augmented Generation (RAG) for use cases like:

- Enterprise Chat
- Question/Answering (QA) over Documents and Code
- AI Workflows / Agents 


## Integrations

###  Large Language Models (LLM)

- OpenAI API (or compatible)  
  - [OpenAI Platform](https://platform.openai.com/docs/introduction)
  - [Azure OpenAI Service](https://azure.microsoft.com/en-us/products/ai-services/openai-service)
  - [vLLM](https://docs.vllm.ai)
  - ...

- Local Models
  - [LLAMA.CPP](https://github.com/ggerganov/llama.cpp)
  - [Ollama](https://ollama.ai/)
  - [Sentence BERT](https://www.sbert.net) 


### Vector Indexes

- [Chroma](https://www.trychroma.com) Embedding Database
- [Weaviate](https://weaviate.io) Vector Database
- In-Memory cosine similarity

### User Authorizers

- Static Token
- OIDC JWT Tokens


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
$ server --port 9081 --model ./models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
```

```shell
$ docker run -it --rm -p 9081:9081 -v ./models/:/models/ ghcr.io/ggerganov/llama.cpp:full --server --host 0.0.0.0 --port 9081 --model /models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
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

#### Ollama

```shell
$ ollama start
$ ollama run mistral
```

```yaml
providers:
  - type: ollama
    url: http://localhost:11434

    models:
      mistral-7b-instruct:
        id: mistral
```

#### Sentence-BERT

```shell
$ docker run -it --rm -p 9082:8080 semitechnologies/transformers-inference:sentence-transformers-all-mpnet-base-v2
```

```yaml
providers:
  - type: sbert
    url: http://localhost:9082

    models:
      all-mpnet-base-v2:
        id: all-mpnet-base-v2
```

### Indexes

#### Chroma

```shell
$ docker run -it --rm -p 9083:8000 ghcr.io/chroma-core/chroma
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
$ docker run -it --rm -p 9084:8080 -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true -e PERSISTENCE_DATA_PATH=/data semitechnologies/weaviate
```

```yaml
indexes:
  docs:
    type: weaviate
    url: http://localhost:9084
    namespace: Document
    embedding: text-embedding-ada-002
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

#### Configuration

```yaml
chains:
  qa:
    type: rag
    index: docs
    model: mistral-7b-instruct
    limit: 5
    distance: 0.5
```

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

### Function Calling Mimicking

For providers or models not natively supporting Function Calling, a transformator chain can be configured to mimic this functionality

#### Configuration

```yaml
chains:
  fn:
    type: fn
    model: mistral-7b-instruct
```