#!/usr/bin/env python
import os
import uvicorn

from typing import Any, List, Union
from fastapi import FastAPI
from langserve import add_routes

from langchain import hub
from langchain_openai import ChatOpenAI
from langchain.agents import AgentExecutor, create_react_agent
from langchain_core.messages import AIMessage, FunctionMessage, HumanMessage

from langchain.pydantic_v1 import BaseModel, Field
from langchain_community.tools import DuckDuckGoSearchResults

llm = ChatOpenAI(model_name=os.environ['MODEL_NAME'], temperature=0, streaming=True)

# https://smith.langchain.com/hub/hwchase17/react-chat
prompt = hub.pull("hwchase17/react-chat")

tools = [DuckDuckGoSearchResults(max_results=1)]
agent = create_react_agent(llm, tools, prompt)

runnable = AgentExecutor(agent=agent, tools=tools, verbose=True, handle_parsing_errors=True)

app = FastAPI(title="LangChain Server")

class Input(BaseModel):
    input: str

    chat_history: List[Union[HumanMessage, AIMessage, FunctionMessage]] = Field(
        ...,
        extra={"widget": {"type": "chat", "input": "input", "output": "output"}},
    )

class Output(BaseModel):
    output: Any

add_routes(
    app,
    runnable.with_types(input_type=Input, output_type=Output).with_config(
        {"run_name": "agent"}
    ),
)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)