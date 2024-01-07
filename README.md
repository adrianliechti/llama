
# LLAMA Platform

The LLAMA Platform acts as a comprehensive gateway, offering a streamlined API that grants access to sophisticated Large Language Models (LLMs). It simplifies the process of integrating advanced natural language processing capabilities into various applications and services.

Beyond providing basic access to these models, LLAMA Platform goes a step further by including higher-level functions such as Retrieval-Augmented Generation (RAG), which significantly enhances the quality and relevance of generated outputs by first sourcing related information from expansive datasets.

Through LLAMA Platform's unified interface, developers can effortlessly harness the power of these AI models without delving into the complexities of underlying machine learning frameworks. This opens up the potential for widespread adoption of AI capabilities across numerous industries, driving innovation and efficiency at scale.

## Use Cases

With its versatile framework, the LLAMA Platform is particularly well-suited to facilitate a variety of use cases, including but not limited to:

- Enterprise Chat: Businesses can deploy sophisticated natural language chatbots that are capable of understanding and responding to complex inquiries with high accuracy and context relevance. These chatbots can serve in customer service, HR support, tech assistance, and many other corporate functions, greatly enhancing the user experience while streamlining operations.

- Question/Answering (QA) over Documents and Code: The platform's advanced search and retrieval capabilities allow for high-precision QA over extensive collections of corporate documents, legal texts, or vast codebases. Users can rapidly find answers to their questions, understand document contents, or fix and improve code, making information accessible even in highly specialized domains.

- AI Workflows / Agents: LLAMA Platform can facilitate the creation of AI-powered workflow automation. This encompasses intelligent agents that predict and execute next steps in a workflow, interact with other digital systems to gather information or perform tasks, and assist human operators with decision-making processes in complex scenarios.

In addition, the LLAMA Platform can be tailored to support more specific applications such as:

- Content Creation and Summarization: Media and publishing agencies can leverage LLMs for rapidly generating high-quality written content, or for condensing long-form articles into concise summaries without losing the essence of the original texts.

- Language Translation and Localization: LLAMA Platform's access to multilingual LLMs allows for real-time translation services and localization of products or services, making them accessible to a wider, global audience.

- Personalized Recommendations: E-commerce platforms and content providers can use RAG-based systems to craft highly personalized product or content recommendations by analyzing user queries and behavior with unprecedented depth.

- Learning and Education: Educational platforms can implement the RAG feature for creating adaptive learning materials that provide tailored explanations, generate practice questions, and interactively guide students through complex subjects.


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

#### Sentence-BERT

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


### Vector Databses / Indexes

Indexes / Vector databases are specialized storage systems designed to efficiently handle vector data, which typically arise from embeddings generated by machine learning models, particularly in the domains of natural language processing (NLP) and computer vision.

Embeddings are high-dimensional vectors that represent complex data like text, images, or audio in a numerical form that captures the inherent properties of the input data. The distance or angle between these vectors can be used to measure the similarity or dissimilarity of the original data points.


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


### Classifications

Use classifiers to categorize prompts and dispatch or filter data in chains

- Sentiment Analysis: Determining whether a piece of text expresses a positive, negative, or neutral sentiment. This is useful for businesses looking to monitor brand perception and customer satisfaction.
- Content Categorization: Automatically tagging articles, blog posts, or documents with relevant categories (e.g., sports, politics, technology), which assists in organizing and recommending content in digital platforms.
- Spam Detection: Identifying and filtering out spam emails or comments on websites by understanding the common characteristics of spammy content.
- Intent Recognition: Determining the intent behind user queries in chatbots or virtual assistants, such as whether a user is asking a question, making a request, or expressing a complaint.
- Toxicity Moderation: Automatically detecting and flagging offensive or inappropriate language in online platforms, helping maintain community standards and a positive user environment.
- Language Identification: Determining which language a given piece of text is written in, which is especially useful for multilingual platforms.

#### LLM Classifier

```yaml
classifiers:
  {classifier-id}:
    type: llm
    model: {model-id}
    categories:
      {category-1}: "...Description when to use Category 1..."
      {category-2}: "...Description when to use Category 2..."
``````

## Use Cases

### Retrieval Augmented Generation (RAG)

Retrieval-Augmented Generation represents a paradigm where a generative model is bolstered by an external retrieval mechanism during the text generation process.

- Retrieval: First, when prompted with input (such as a question or a prompt for continuation), the system retrieves relevant pieces of information from a large dataset or database (e.g. vector database). This information might consist of facts, quotes, passages, or any other form of textual data that is relevant to the input.

- Generation: After the relevant information has been retrieved, a generative model incorporates this information to generate a coherent and contextually appropriate response.

The advantages of Retrieval-Augmented Generation include:

- Augmented Knowledge: Generative models can benefit from access to a broad range of knowledge not contained within their parameters, essentially using the external database to "look up" information relevant to the task at hand.

- Improved Factuality: By grounding the generated text in retrieved documents that are assumed to be factual, the outputs of the model can be more accurate and less prone to making unsupported statements.

- Flexibility: The retrieval database can be updated with new information without needing to retrain the generative model, making the system more adaptable to new data and domains.

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

Function calling is the ability to reliably connect LLMs to external tools to enable effective tool usage and interaction with external APIs. Function calling is an important ability for building LLM-powered chatbots or agents that need to retrieve context for an LLM or interact with external tools by converting natural language into API calls.

- conversational agents that can efficiently use external tools to answer questions. For example, the query "What is the weather like in Belize?" will be converted to a function call such as get_current_weather(location: string, unit: 'celsius' | 'fahrenheit')
- LLM-powered solutions for extracting and tagging data (e.g., extracting people names from a Wikipedia article)
- applications that can help convert natural language to API calls or valid database queries
- conversational knowledge retrieval engines that interact with a knowledge base

Example using OpenAI API https://platform.openai.com/docs/guides/function-calling

#### Compatibility Layer

For providers or models not natively supporting Function Calling, a transformator chain can be configured to mimic this functionality.

#### Configuration

```yaml
chains:
  fn:
    type: fn
    model: mistral-7b-instruct
```