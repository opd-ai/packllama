# packllama API Examples

This page shows quick-start `curl` commands for every packllama API endpoint.
All examples assume the server is running on `http://127.0.0.1:8080`.

---

## Health Check

```bash
curl http://127.0.0.1:8080/health
```

**Response**

```json
{"status":"ok"}
```

---

## Chat Completions (`/v1/chat/completions`)

### Non-streaming

```bash
curl -s http://127.0.0.1:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3-8b-instruct",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user",   "content": "Explain Go interfaces in one paragraph."}
    ],
    "temperature": 0.7,
    "max_tokens": 256
  }'
```

**Response**

```json
{
  "id": "a1b2c3d4...",
  "object": "chat.completion",
  "created": 1718000000,
  "model": "llama-3-8b-instruct",
  "choices": [
    {
      "index": 0,
      "message": {"role": "assistant", "content": "In Go, an interface is a type..."},
      "finish_reason": "stop"
    }
  ],
  "usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
}
```

### Streaming (Server-Sent Events)

```bash
curl -s http://127.0.0.1:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3-8b-instruct",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

Each line of the response is a `data:` SSE frame containing a JSON chunk, ending with `data: [DONE]`.

---

## Text Completions (`/v1/completions`)

### Non-streaming

```bash
curl -s http://127.0.0.1:8080/v1/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "codellama-13b",
    "prompt": "func add(a, b int) int {",
    "max_tokens": 64,
    "stop": ["}"]
  }'
```

### Streaming

```bash
curl -s http://127.0.0.1:8080/v1/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "codellama-13b",
    "prompt": "The answer to life is",
    "max_tokens": 20,
    "stream": true
  }'
```

---

## Models (`/v1/models`)

### List all models

```bash
curl -s http://127.0.0.1:8080/v1/models
```

**Response**

```json
{
  "object": "list",
  "data": [
    {
      "id": "llama-3-8b-instruct",
      "object": "model",
      "created": 1718000000,
      "owned_by": "local",
      "context_length": 8192,
      "parameter_count": 8000000000,
      "quantization": "Q4_K_M"
    }
  ]
}
```

> **Note**: `context_length`, `parameter_count`, and `quantization` are populated once an
> inference backend is configured. They are omitted when unknown.

### Get a single model

```bash
curl -s http://127.0.0.1:8080/v1/models/llama-3-8b-instruct
```

### Load a model

```bash
curl -s http://127.0.0.1:8080/v1/models \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/opt/models/llama-3-8b-instruct.gguf",
    "id": "llama-3-8b-instruct"
  }'
```

### Unload a model

```bash
curl -s -X DELETE http://127.0.0.1:8080/v1/models/llama-3-8b-instruct
```

---

## Embeddings (`/v1/embeddings`)

### Single input

```bash
curl -s http://127.0.0.1:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "nomic-embed-text",
    "input": "The quick brown fox"
  }'
```

### Batch input

```bash
curl -s http://127.0.0.1:8080/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{
    "model": "nomic-embed-text",
    "input": ["The quick brown fox", "jumped over the lazy dog"]
  }'
```

**Response**

```json
{
  "object": "list",
  "data": [
    {"object": "embedding", "index": 0, "embedding": [0.023, -0.045, ...]},
    {"object": "embedding", "index": 1, "embedding": [0.011, -0.032, ...]}
  ],
  "model": "nomic-embed-text",
  "usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
}
```

---

## Error Responses

All error responses share the same JSON shape:

```json
{"error": "description of the problem"}
```

| HTTP status | When it occurs |
|-------------|----------------|
| 400 | Request body is not valid JSON |
| 422 | Required field missing or invalid value |
| 404 | Model ID not found |
| 500 | Inference backend returned an error |
| 503 | Inference service not configured |

---

## Using with the OpenAI Go SDK

```go
import "github.com/openai/openai-go"

client := openai.NewClient(
    option.WithBaseURL("http://127.0.0.1:8080/v1"),
    option.WithAPIKey("not-required"),
)
```

## Using with the OpenAI Python SDK

```python
from openai import OpenAI

client = OpenAI(base_url="http://127.0.0.1:8080/v1", api_key="not-required")
response = client.chat.completions.create(
    model="llama-3-8b-instruct",
    messages=[{"role": "user", "content": "Hello!"}],
)
```
