import os
import streamlit as st
from openai import OpenAI

client = OpenAI(
    base_url=os.getenv("OPENAI_BASE_URL", "http://localhost:8080/oai/v1/"),
    api_key=os.getenv("OPENAI_API_KEY", "-"))

st.set_page_config(
    page_title="LLM Platform Chat"
)

@st.cache_resource
def get_models():
    return list(filter(
       lambda m: "embed" not in m.id and "tts" not in m.id and "whisper" not in m.id,
       sorted(client.models.list(), key=lambda m: m.id)
    ))

with st.sidebar:
    st.title("LLM Platform Chat")
    
    st.selectbox("Model", get_models(), key="model", format_func=lambda m: m.id)
    st.text_area("System Prompt", "", key="system")
    st.slider(label="Temperature", key="temperature", min_value=0.0, max_value=1.0, value=0.0, step=.1)

if "messages" not in st.session_state:
    st.session_state.messages = []

for message in st.session_state.messages:
    with st.chat_message(message["role"]):
        st.markdown(message["content"])

if prompt := st.chat_input("What is up?"):
    st.session_state.messages.append({"role": "user", "content": prompt})

    with st.chat_message("user"):
        st.markdown(prompt)

    with st.chat_message("assistant"):
        messages = []

        if st.session_state.system:
            messages.append({"role": "system", "content": st.session_state.system})
        
        for m in st.session_state.messages:
            messages.append({"role": m["role"], "content": m["content"]})

        stream = client.chat.completions.create(
            model=st.session_state.model.id,
            messages=messages,
            temperature=st.session_state.temperature,
            stream=True,
        )

        response = st.write_stream(stream)
    
    st.session_state.messages.append({"role": "assistant", "content": response})