FROM python:3-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD [ "streamlit", "run", "app.py", "--server.address=0.0.0.0", "--server.port=8501", "--client.toolbarMode=viewer" ]
