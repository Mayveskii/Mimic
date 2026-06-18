# LlamaIndex Setup

Route LlamaIndex OpenAI LLM calls through GonkaGate.

Route LlamaIndex OpenAI LLM calls through GonkaGate by setting `api_base`, a `gp-...` API key, and a current GonkaGate model ID. Keep the rest of your LlamaIndex pipeline in place. This guide covers the OpenAI LLM wrapper only. Keep embeddings and broader RAG wiring on your current provider or local model.

## Configure `Settings.llm`

Use `Settings.llm` when the same GonkaGate-backed LLM should be reused across multiple query or index flows.
 Configure Settings.llm
```
from llama_index.core import Settings
from llama_index.llms.openai import OpenAI

Settings.llm = OpenAI(
    model="qwen/qwen3-235b-a22b-instruct-2507-fp8",
    api_key="gp-your-api-key",
    api_base="https://api.gonkagate.com/v1",
)

response = Settings.llm.complete("Return exactly: LlamaIndex connected")
print(response)
```

Expected result: `Settings.llm.complete(...)` returns `LlamaIndex connected`.

Use a fresh model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models) before you send real traffic. If your code already instantiates `OpenAI(...)` locally, apply the same `api_base`, API key, and model ID there.

## Pass the same LLM into your query flow

If your app already has an index, pass the configured LLM into `as_query_engine()` instead of rebuilding the pipeline.
 Pass the same LLM into your query flow
```
query_engine = index.as_query_engine(llm=Settings.llm)
response = query_engine.query("Summarize this document in one sentence.")
print(response)
```

## Change only the LLM connection values

- Keep the LlamaIndex OpenAI wrapper in place.
- Change only `api_base`, the API key, and the model ID.
- Use `Settings.llm` when several flows should share the same LLM configuration.
- Revisit retrieval or broader RAG design only after the base LLM connection works.

## Common failures

| Response or symptom | What it usually means | What to do |
|---|---|---|
| 401 invalid_api_key | The API key is missing, invalid, or loaded from the wrong place | Recheck Authentication and API Keys |
| 404 model_not_found | The model ID is stale or unsupported | Refresh it from GET /v1/models |
| 429 insufficient_quota | The prepaid USD balance is too low for the request | Top up the balance, then retry after funds are available |
| 429 rate_limit_exceeded | You hit a runtime limit | Honor Retry-After and add bounded backoff |
| LLM calls work but embeddings fail | This guide covers only the OpenAI LLM wrapper path | Keep embeddings on your current provider or local model |

## See also

- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) for a broader OpenAI-compatible switch
- [Chat Completions API reference](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact request and response contract behind the LLM wrapper
- [GonkaGate API Error Handling](/en/docs/api/reference/error-handling) for retry policy and failure handling after the base connection works