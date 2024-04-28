# Kubernetes Example

## Models
- Meta Llama 3 8B
- Microsoft Phi 3 Mini

## Installation

```bash
kubectl create namespace llm-demo
kubectl apply -n llm-demo -k https://github.com/adrianliechti/llama/examples/kubernetes-chat/
kubectl port-forward service/chat 8501:80 -n llm-demo
```

## Demo Client

```bash
kubectl port-forward service/chat :80 -n llm-demo
```

Open http://localhost:8501 in your favorite browser