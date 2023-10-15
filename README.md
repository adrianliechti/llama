
# Llama Server

https://github.com/ggerganov/llama.cpp


### Download Code Lama GGUF Model

```bash
mkdir ./models
curl -L -o ./models/llama-2-7b-chat.Q4_K_M.gguf https://huggingface.co/TheBloke/Llama-2-7b-Chat-GGUF/resolve/main/llama-2-7b-chat.Q4_K_M.gguf
```

### Run Llama Server

```bash
docker pull ghcr.io/ggerganov/llama.cpp:full
docker run -it --rm -p 8000:8000 -v $(pwd)/models:/modelsghcr.io/ggerganov/llama.cpp:full --server --host 0.0.0.0 --port 8000 --path /public --model /models/llama-2-7b-chat.Q4_K_M.gguf --embedding --alias default
```

### Run OpenAI API Server

```bash
docker pull adrianliechti/llama-openai
docker run -it --rm -p 8080:8080 -e LLAMA_URL=http://host.docker.internal:8000 adrianliechti/llama-openai
```

### Run ChatBot UI

```bash
docker pull ghcr.io/mckaywrigley/chatbot-ui:main
docker run -it --rm -p 3000:3000 -e OPENAI_API_HOST=http://host.docker.internal:8080 -e OPENAI_API_KEY=none -e DEFAULT_MODEL=default ghcr.io/mckaywrigley/chatbot-ui:main
```