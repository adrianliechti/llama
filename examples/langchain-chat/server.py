#!/usr/bin/env python
import os
import uvicorn

from typing import Any
from fastapi import FastAPI
from langserve import add_routes

from langchain import hub
from langchain_openai import ChatOpenAI
from langchain.agents import AgentExecutor, create_react_agent

from langchain.pydantic_v1 import BaseModel
from langchain_community.tools import DuckDuckGoSearchResults

llm = ChatOpenAI(model_name=os.environ['MODEL_NAME'], streaming=True)

prompt = hub.pull("hwchase17/react")

tools = [DuckDuckGoSearchResults(max_results=1)]
agent = create_react_agent(llm, tools, prompt)

runnable = AgentExecutor(agent=agent, tools=tools, verbose=True)

app = FastAPI(title="LangChain Server")

class Input(BaseModel):
    input: str

class Output(BaseModel):
    output: Any

add_routes(app, runnable.with_types(input_type=Input, output_type=Output))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)