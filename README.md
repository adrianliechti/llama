
# LLM Platform

<img src="docs/icon.png" width="150"/>

The LLM Platform or Inference Hub is an open-source product designed to simplify the development and deployment of large language model (LLM) applications at scale. It provides a unified framework that allows developers to integrate and manage multiple LLM vendors, models, and related services through a standardized but highly flexible approach.

## Key Features

### Multi-Provider Support

The platform integrates with a wide range of LLM providers, including but not limited to

- OpenAI Platform and Azure OpenAI Service to access models such as GPT, DALL-E and Whisper
- Anthropic, Cohere, ElevenLabs, Google, Groq, Jina, Mistral and Replicate for various specialised models.
- Local deployments such as Ollama, LLAMA.CPP, WHISPER.CPP and Mistral.RS for running models locally.
- Community models via Hugging Face
- Custom models via gRPC plugins

### Flexible Configuration

Developers can define providers, models, credentials, vector databases, tools, document extractors or advanced chains using YAML configuration files. This approach streamlines the integration process and makes it easier to manage multiple services and models.

### Routing and Load Balancing

The platform includes routing capabilities such as a round-robin load balancer to efficiently distribute requests across multiple models or providers. This increases scalability and ensures high availability.

### Vector Databases and Indexes

Supports integration with various vector databases and indexing services for efficient data retrieval and storage.

Supported systems include
- SaaS offerings such as Azure Search
- Self-hosting solutions such as ChromaDB, Qdrant, Weaviate, Postgres or Elasticsearch
- Custom retrievers via gRPC plugins
- In-memory and temporary indexes

### Observability

The platform is fully traceable using OpenTelemetry, which provides comprehensive observability and monitoring of the entire system and its components. This increases transparency and reliability, enabling proactive maintenance and smoother operation of LLM applications at scale.


## Architecture

![Architecture](docs/architecture.png)

The architecture is designed to be modular and extensible, allowing developers to plug in different providers and services as needed. It consists of a number of key components:

- Providers: Interface to various AI / LLM services.
- Indexes: Handle data storage and retrieval
- Extractors: Process and extract data from documents or web pages
- Segmenters: Semantically split text into chunks for RAG
- Summarisers: Compress large texts or prompts
- Translate: Translate prompt input or output or entire documents
- Routers & Rate Limiters: Manage how requests are distributed across models
- Tools: Pre-built or custom tools for translating, retrieving documents or searching the web using function calls.

## Use Cases:

- Unified enterprise chat using multiple sources and specialised agents
- Scalable LLM applications: Ideal for building applications that need to scale horizontally and handle high volumes of requests
- Multi-model deployment: Useful for applications that require access to different models from different vendors
- Custom workflows: Enables the creation of custom NLP workflows by combining different services and models


## Integrations & Configuration

### LLM Providers

#### OpenAI Platform

https://platform.openai.com/docs/api-reference

```yaml
providers:
  - type: openai
    token: sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

    models:
      - gpt-4o
      - gpt-4o-mini
      - text-embedding-3-small
      - text-embedding-3-large
      - whisper-1
      - dall-e-3
      - tts-1
      - tts-1-hd
```


#### Azure OpenAI Service

https://azure.microsoft.com/en-us/products/ai-services/openai-service

```yaml
providers:
  - type: openai
    url: https://xxxxxxxx.openai.azure.com
    token: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

    models:
      # https://docs.anthropic.com/en/docs/models-overview
      #
      # {alias}:
      #   - id: {azure oai deployment name}

      gpt-3.5-turbo:
        id: gpt-35-turbo-16k

      gpt-4:
        id: gpt-4-32k
        
      text-embedding-ada-002:
        id: text-embedding-ada-002
```


#### Anthropic

https://www.anthropic.com/api

```yaml
providers:
  - type: anthropic
    token: sk-ant-apixx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

    # https://docs.anthropic.com/en/docs/models-overview
    #
    # {alias}:
    #   - id: {anthropic api model name}
    models:
      claude-3.5-sonnet:
        id: claude-3-5-sonnet-20240620
```


#### Cohere

```yaml
providers:
  - type: cohere
    token: ${COHERE_API_KEY}

    # https://docs.cohere.com/docs/models
    #
    # {alias}:
    #   - id: {cohere api model name}
    models:
      cohere-command-r-plus:
        id: command-r-plus
      
      cohere-embed-multilingual-v3:
        id: embed-multilingual-v3.0
```


#### Groq

```yaml
providers:
  - type: groq
    token: ${GROQ_API_KEY}

    # https://console.groq.com/docs/models
    #
    # {alias}:
    #   - id: {groq api model name}
    models:
      groq-llama-3-8b:
        id: llama3-8b-8192

      groq-whisper-1:
        id: whisper-large-v3
```


#### Mistral AI

```yaml
providers:
  - type: mistral
    token: ${MISTRAL_API_KEY}

    # https://docs.mistral.ai/getting-started/models/
    #
    # {alias}:
    #   - id: {mistral api model name}
    models:
      mistral-large:
        id: mistral-large-latest
```


#### Replicate

https://replicate.com/

```yaml
providers:
  - type: replicate
    token: ${REPLICATE_API_KEY}
    #
    # {alias}:
    #   - id: {cohere api model name}
    models:
      replicate-flux-pro:
        id: black-forest-labs/flux-pro
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

    # https://ollama.com/library
    #
    # {alias}:
    #   - id: {ollama model name with optional version}
    models:
      mistral-7b-instruct:
        id: mistral:latest
```


#### LLAMA.CPP

https://github.com/ggerganov/llama.cpp/tree/master/examples/server

```shell
# using taskfile.dev
$ task llama:server

# LLAMA.CPP Server
$ llama-server --port 9081 --log-disable --model ./models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
```

```yaml
providers:
  - type: llama
    url: http://localhost:9081

    models:
      - mistral-7b-instruct
```


#### Mistral.RS

https://github.com/EricLBuehler/mistral.rs

```shell
$ mistralrs-server --port 1234 --isq Q4K plain -m meta-llama/Meta-Llama-3.1-8B-Instruct -a llama
```

```yaml
providers:
  - type: mistralrs
    url: http://localhost:1234

    models:
      mistralrs-llama-3.1-8b:
        id: llama
        
```


#### WHISPER.CPP

https://github.com/ggerganov/whisper.cpp/tree/master/examples/server

```shell
# using taskfile.dev
$ task whisper:server

# WHISPER.CPP Server
$ whisper-server --port 9083 --convert --model ./models/whisper-large-v3-turbo.bin
```

```yaml
providers:
  - type: whisper
    url: http://localhost:9083

    models:
      - whisper
```


#### Hugging Face

https://huggingface.co/

```yaml
providers:
  - type: huggingface
    token: hf_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    
    models:
      mistral-7B-instruct:
        id: mistralai/Mistral-7B-Instruct-v0.1
      
      huggingface-minilm-l6-2:
        id: sentence-transformers/all-MiniLM-L6-v2
```


#### Eleven Labs

```yaml
providers:
  - type: elevenlabs
    token: ${ELEVENLABS_API_KEY}

    models:
      elevenlabs-sarah:
        id: EXAVITQu4vr4xnSDxMaL
      
      elevenlabs-charlie:
        id: IKne3meq5aSn9XLyUdCD
```


### Routers

#### Round-robin Load Balancer

```yaml
routers:
  llama-lb:
    type: roundrobin
    models:
      - llama-3-8b
      - groq-llama-3-8b
      - huggingface-llama-3-8b
```


### Vector Databses / Indexes

#### Chroma

https://www.trychroma.com

```shell
# using Docker
$ docker run -it --rm -p 9083:8000 -v chroma-data:/chroma/chroma ghcr.io/chroma-core/chroma
```

```yaml
indexes:
  docs:
    type: chroma
    url: http://localhost:9083
    namespace: docs
    embedder: text-embedding-3-large
```


#### Weaviate

https://weaviate.io

```shell
# using Docker
$ docker run -it --rm -p 9084:8080 -v weaviate-data:/var/lib/weaviate -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true -e PERSISTENCE_DATA_PATH=/var/lib/weaviate semitechnologies/weaviate
```

```yaml
indexes:
  docs:
    type: weaviate
    url: http://localhost:9084
    namespace: Document
    embedder: text-embedding-3-large
```


#### Qdrant

```shell
$ docker run -p 6333:6333 qdrant/qdrant:v1.11.4
```

```yaml
indexes:
  docs:
    type: qdrant
    url: http://localhost:6333
    namespace: docs
    embedder: text-embedding-3-large
```


#### In-Memory

```yaml
indexes:
  docs:
    type: memory   
    embedder: text-embedding-3-large
```


#### OpenSearch / Elasticsearch

```shell
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


### Extractor

#### Tika

```shell
# using Docker
docker run -it --rm -p 9998:9998 apache/tika:3.0.0.0-BETA2-full
```

```yaml
extractors:  
  tika:
    type: tika
    url: http://localhost:9998
    chunkSize: 4000
    chunkOverlap: 200
```


#### Unstructured

https://unstructured.io

```shell
# using taskfile.dev
task unstructured:server

# using Docker
docker run -it --rm -p 9085:8000 quay.io/unstructured-io/unstructured-api:0.0.80 --port 8000 --host 0.0.0.0
```

```yaml
extractors:
  unstructured:
    type: unstructured
    url: http://localhost:9085/general/v0/general
```