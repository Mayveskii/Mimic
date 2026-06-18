# LangChain Setup

Connect LangChain ChatOpenAI to GonkaGate.

Connect LangChain to GonkaGate by keeping `ChatOpenAI` in place and changing only the API key, base URL, and model ID. Verify one request first. Test tool calling, structured outputs, or larger agent flows only after the base connection works.

## Verify the connection with one request
 PythonTypeScript
```
import os

from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="qwen/qwen3-235b-a22b-instruct-2507-fp8",
    api_key=os.environ["GONKAGATE_API_KEY"],
    base_url="https://api.gonkagate.com/v1",
)

response = llm.invoke("Return exactly: LangChain connected")
print(response.content)
```

If this returns `LangChain connected`, `ChatOpenAI` is reaching GonkaGate. Replace the example model with a fresh ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models) before real traffic.

## Keep the wrapper, change the connection values

- Keep LangChain `ChatOpenAI` in place.
- Change only the API key, base URL, and model ID.
- Keep your existing `invoke()` and `stream()` usage around that wrapper.
- Validate tool calling, structured outputs, or a larger agent flow only after the base connection works.
- LangGraph, embeddings, retrievers, and broader LangChain architecture are outside this guide.

## Common first failures

| If you see | What it usually means | What to do |
|---|---|---|
| 401 invalid_api_key | The API key is wrong or not loaded into the runtime | Check GONKAGATE_API_KEY, secret loading, and key state in Authentication and API Keys |
| 404 model_not_found | The model ID is stale or invalid | Switch to a model ID from GET /v1/models |
| 429 insufficient_quota | The prepaid USD balance is too low for this request | Top up first or reduce request cost before retrying |
| 429 rate_limit_exceeded | The request was throttled | Respect Retry-After and add bounded backoff |
If the base request works but a later LangChain step fails, confirm that step still runs through `ChatOpenAI` before debugging LangChain internals.

## See also

- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) for a broader OpenAI-compatible app migration
- [Create a chat completion](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact request and response contract behind `ChatOpenAI`
- [GonkaGate API Error Handling](/en/docs/api/reference/error-handling) for retry and failure-handling rules after the connection works