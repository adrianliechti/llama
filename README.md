
# LLM Platform

Open Source LLM Platform to build and deploy applications at scale

![Logo](docs/icon.png)


## Architecture

![Architecture](docs/architecture.png)


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

    models:
      # https://docs.anthropic.com/en/docs/models-overview
      #
      # {alias}:
      #   - id: {anthropic api model name}

      claude-3-opus:
        id: claude-3-opus-20240229
```


#### Cohere

```yaml
providers:
  - type: cohere
    token: ${COHERE_API_KEY}

    # https://docs.cohere.com/docs/models
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

    models:
      # https://ollama.com/library
      #
      # {alias}:
      #   - id: {ollama model name with optional version}

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

# LLAMA.CPP Server (Multimodal Model)
$ llama-server --port 9081 --log-disable --model ./models/llava-v1.5-7b-Q4_K.gguf --mmproj ./models/llava-v1.5-7b-mmproj-Q4_0.gguf

# using Docker (might be slow)
$ docker run -it --rm -p 9081:9081 -v ./models/:/models/ ghcr.io/ggerganov/llama.cpp:server --host 0.0.0.0 --port 9081 --model /models/mistral-7b-instruct-v0.2.Q4_K_M.gguf
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
$ whisper-server --port 9083 --convert --model ./models/whisper-ggml-medium.bin
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


#### LangChain / LangServe

https://python.langchain.com/docs/langserve

```yaml
providers:
  - type: langchain
    url: http://your-langchain-server:8000

    models:
      - langchain
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
    embedder: text-embedding-ada-002
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
    embedder: text-embedding-ada-002
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
    embedder: text-embedding-ada-002
```


#### In-Memory

```yaml
indexes:
  docs:
    type: memory   
    embedder: text-embedding-ada-002
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
# using Docker
docker run -it --rm -p 9085:8000 quay.io/unstructured-io/unstructured-api:0.0.75 --port 8000 --host 0.0.0.0
```

```yaml
extractors:
  unstructured:
    type: unstructured
    url: http://localhost:9085
```