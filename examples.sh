#!/bin/bash

export OPENAI_BASE_URL=http://localhost:8080/v1
export OPENAI_API_KEY=changeme
export OPENAI_MODEL=gpt-3.5-turbo
export OPENAI_EMBEDDING_MODEL=text-embedding-ada-002

# Models API

curl -v $OPENAI_BASE_URL/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"|jq

# Embedding API

curl -v $OPENAI_BASE_URL/embeddings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
    "input": "Your text string goes here",
    "model": "'$OPENAI_EMBEDDING_MODEL'"
  }'|jq

# Chat Completion API

curl $OPENAI_BASE_URL/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
    "model": "'$OPENAI_MODEL'",
    "messages": [
      {
        "role": "system",
        "content": "You are a helpful assistant."
      },
      {
        "role": "user",
        "content": "Who won the world series in 2020?"
      },
      {
        "role": "assistant",
        "content": "The Los Angeles Dodgers won the World Series in 2020."
      },
      {
        "role": "user",
        "content": "Where was it played?"
      }
    ]
  }'|jq

curl -v -N $OPENAI_BASE_URL/chat/completions \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $OPENAI_API_KEY" \
-d '{
  "stream": true,
  "model": "'$OPENAI_MODEL'",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Who won the world series in 2020?"
    },
    {
      "role": "assistant",
      "content": "The Los Angeles Dodgers won the World Series in 2020."
    },
    {
      "role": "user",
      "content": "Where was it played?"
    }
  ]
}'