
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
    token: ${COHERE_API_TOKEN}

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
    token: ${GROQ_API_TOKEN}

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
    token: ${MISTRAL_API_TOKEN}

    # https://docs.mistral.ai/getting-started/models/
    models:
      mistral-large:
        id: mistral-large-latest
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
$ mistralrs-server --port 1234 vision-plain -m microsoft/Phi-3-vision-128k-instruct -a phi3v
```

```yaml
providers:
  - type: mistralrs
    url: http://localhost:1234

    models:
      - phi-3-vision
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
    url: http://localhost:9085

    models:
      - whisper
```


#### Hugging Face

https://huggingface.co/

```yaml
providers:
  - type: huggingface
    url: https://api-inference.huggingface.co/models/mistralai/Mistral-7B-Instruct-v0.1
    token: hf_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    
    models:
      - mistral-7B-instruct
  
  - type: huggingface
    url: https://api-inference.huggingface.co/pipeline/feature-extraction/sentence-transformers/all-MiniLM-L6-v2
    token: hf_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

    models:
      - huggingface-minilm-l6-2
```


#### Mimic 3

```shell
mkdir -p mimic3
chmod 777 mimic3
docker run -it -p 59125:59125 -v $(pwd)/models/mimic3:/home/mimic3/.local/share/mycroft/mimic3 mycroftai/mimic3
```

```yaml
providers:
  - type: mimic
    url: http://localhost:59125

    models:
      - mimic-3
```

```shell
docker run --rm -it -p 5002:5002 --platform linux/amd64 --entrypoint /bin/bash ghcr.io/coqui-ai/tts-cpu
python3 TTS/server/server.py --list_models
python3 TTS/server/server.py --model_name tts_models/en/vctk/vits
```


#### Coqui

```yaml
providers:
  - type: coqui
    url: http://localhost:5002

    models:
      - coqui-1
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
    embedding: text-embedding-ada-002
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
    embedding: text-embedding-ada-002
```


#### Qdrant

```shell
$ docker run -p 6333:6333 qdrant/qdrant:v1.10.1
```

```yaml
indexes:
  docs:
    type: qdrant
    url: http://localhost:6333
    namespace: docs
    embedding: text-embedding-ada-002
```


#### In-Memory

```yaml
indexes:
  docs:
    type: memory   
    embedding: text-embedding-ada-002
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


### Extractors

#### Text

```yaml
extractors:
  text:
    type: text
```


#### Code

```yaml
extractors:
  code:
    type: code
```


#### Tesseract

https://tesseract-ocr.github.io

```shell
# using Docker
docker run -it --rm -p 9086:8884 hertzg/tesseract-server:latest
```

```yaml
extractors:
  tesseract:
    type: tesseract
    url: http://localhost:9086
```


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
```


#### Index Documents

Using Extractor

```
POST http://localhost:8080/v1/index/{index-name}/{extractor}
Content-Type: application/pdf
Content-Disposition: attachment; filename="filename.pdf"
```

Using Documents

```
POST http://localhost:8080/v1/index/{index-name}
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