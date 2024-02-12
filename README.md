
# LLAMA Platform

Open Source LLM Platform to build and deploy applications at scale

## Integrations & Configuration

### LLM Providers

#### OpenAI Platform

https://platform.openai.com/docs/api-reference

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

https://azure.microsoft.com/en-us/products/ai-services/openai-service

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


#### Ollama

https://ollama.ai

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


#### LLAMA.CPP

https://github.com/ggerganov/llama.cpp/tree/master/examples/server

```shell
# using taskfile.dev
$ task llama-server

# LLAMA.CPP Server
$ bin/llama-server --port 9081 --log-disable --embedding --model ./models/mistral-7b-instruct-v0.2.Q4_K_M.gguf

# LLAMA.CPP Server (Multimodal Model)
$ bin/llama-server --port 9081 --log-disable --embedding --model ./models/llava-v1.5-7b-Q4_K.gguf --mmproj ./models/llava-v1.5-7b-mmproj-Q4_0.gguf

# using Docker (might be slow)
$ docker run -it --rm -p 9081:9081 -v ./models/:/models/ ghcr.io/ggerganov/llama.cpp:server --host 0.0.0.0 --port 9081 --model /models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
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

#### WHISPER.CPP

https://github.com/ggerganov/whisper.cpp/tree/master/examples/server

```shell
# using taskfile.dev
$ task whisper-server
```

```yaml
providers:
  - type: whisper
    url: http://localhost:9085

    models:
      whisper:
        id: whisper
```

#### Sentence-BERT (text2vec-transformers)

https://www.sbert.net  
https://github.com/weaviate/t2v-transformers-models

```shell
# using taskfile.dev
task sbert-server

# using Docker
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


#### LangChain / LangServe

https://python.langchain.com/docs/langserve

```yaml
providers:
  - type: langchain
    url: http://your-langchain-server:8000

    models:
      langchain:
        id: default
```


### Vector Databses / Indexes

#### Chroma

https://www.trychroma.com

```shell
# using taskfile.dev
$ task chroma-server

# using Docker
$ docker run -it --rm -p 9083:8000 -v chroma-data:/chroma/chroma ghcr.io/chroma-core/chroma
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

https://weaviate.io

```shell
# using taskfile.dev
$ task weaviate-server

# using Docker
$ docker run -it --rm -p 9084:8080 -v weaviate-data:/var/lib/weaviate -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true -e PERSISTENCE_DATA_PATH=/var/lib/weaviate semitechnologies/weaviate
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


##### OpenSearch / Elasticsearch

```shell
# using taskfile.dev
$ task opensearch-server

# using Docker
docker run -it --rm -p 9200:9200 -v opensearch-data:/usr/share/opensearch/data -e "discovery.type=single-node" -e DISABLE_SECURITY_PLUGIN=true opensearchproject/opensearch:latest
```

```yaml
indexes:
  docs:
    type: elasticsearch
    url: http://localhost:9200
    namespace: docs
```

### Extracters

#### Tesseract

https://tesseract-ocr.github.io

```shell
# using taskfile.dev
$ task tesseract-server

# using Docker
docker run -it --rm -p 9086:8884 hertzg/tesseract-server:latest
```

```yaml
extracters:
  tesseract:
    type: tesseract
    url: http://localhost:9086
```

#### Unstructured

https://unstructured.io

```shell
# using taskfile.dev
$ task unstructured-server

# using Docker
docker run -it --rm -p 9085:8000 downloads.unstructured.io/unstructured-io/unstructured-api:latest --port 8000 --host 0.0.0.0
```

```yaml
extracters:
  unstructured:
    type: unstructured
    url: http://localhost:9085
```


### Classifications

#### LLM Classifier

```yaml
classifiers:
  {classifier-id}:
    type: llm
    model: mistral-7b-instruct
    classes:
      class-1: "...Description when to use Class 1..."
      class-2: "...Description when to use Class 2..."
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

    # limit: 10
    # distance: 1

    # filters:
    #  {metadata-key}:
    #    classifier: {classifier-id}
```

#### Index Documents

Using Extractor

```
POST http://localhost:8080/api/index/{index-name}/{extractor}
Content-Type: application/pdf
Content-Disposition: attachment; filename="filename.pdf"
```

Using Documents

```
POST http://localhost:8080/api/index/{index-name}
```

```json
[
    {
        "id": "id1",
        "content": "content of document...",
        "metadata": {
          "key1": "value1",
          "key2": "value2"
        }
    },
    {
        "id": "id2",
        "content": "content of document...",
        "metadata": {
          "key1": "value1",
          "key2": "value2"
        }
    }
]
```

### Function Calling

#### ReAct

For providers or models not natively supporting Function Calling, a transformator chain can be configured to mimic this functionality.

```yaml
chains:
  mistral-7b-react:
    type: react
    model: mistral-7b-instruct
```