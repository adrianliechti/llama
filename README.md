
# LLAMA Gateway

The LLAMA Gateway provides an unified API to various Large Language Models (LLM) and higher level functionality like 
Retrieval-Augmented Generation (RAG) for use cases like:

- Enterprise Chat
- Question/Answering (QA) over Documents and Code


## Integrations

### Language Models

- OpenAI API (or compatible)  
  (e.g [OpenAI Platform](https://platform.openai.com/docs/introduction), [Azure OpenAI Service](https://azure.microsoft.com/en-us/products/ai-services/openai-service), [vLLM](https://docs.vllm.ai), ...))
- Local & Open Source Models via [llama.cpp](https://github.com/ggerganov/llama.cpp) Server
- Embedding Models using [Sentence Transformers](https://www.sbert.net) 


### Vector Stores / Indexes

- [Weaviate](https://weaviate.io) Vector Database
- [Chroma](https://www.trychroma.com) Embedding Database

### Authorizers

- Static Token
- JWT using OIDC


## Example Application

The Docker `compose.yaml` file starts a simple web-based [Chatbot UI](https://github.com/mckaywrigley/chatbot-ui)  (port http/3000) and a [llama.cpp](https://github.com/ggerganov/llama.cpp) server. The LLAMA Gateway API is exposed locally (port http/8080) using the static token `changeme`. An OpenAI-compatible API is availabe in http://localhost:8080/oai/v1. 

While starting up, a [Mistral 7B Instruct](https://mistral.ai/news/announcing-mistral-7b/) model file will be downloaded from [Hugging Face](https://huggingface.co) (see ./models) if not already exists.

The sample also starts a text2vec sentence-transformer provided by [Weaviate](https://weaviate.io/developers/weaviate/modules/retriever-vectorizer-modules/text2vec-transformers) to provide embedding APIs.

For broad compatibility with existing tools (like the bundled WebUI), the models are aliased as `gpt-3.5-turbo` and `text-embedding-ada-002`.

Run example application

```bash
$ docker compose up
```

Browse to http://localhost:3000