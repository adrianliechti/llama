#!/usr/bin/env python
"""Example LangChain server exposes multiple runnables (LLMs in this case)."""

import os

from fastapi import FastAPI
from langchain_openai import ChatOpenAI

from langserve import add_routes

app = FastAPI(
    title="LangChain Server",
    version="1.0",
    description="Spin up a simple api server using Langchain's Runnable interfaces",
)

add_routes(
    app,
    ChatOpenAI(model_name=os.environ['MODEL_NAME']),
    path="/default",
)

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)