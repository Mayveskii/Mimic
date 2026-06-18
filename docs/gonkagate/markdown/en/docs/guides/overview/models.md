# Model Selection

Choose a GonkaGate model ID for API requests.

Use `GET /v1/models` to choose a GonkaGate model ID for application code and keep `model` configurable. Use `/models` to browse the live catalog and `/pricing` when the question is billing rather than model selection.

## Choose a current model ID

Start with a current model ID from `GET /v1/models`:
 Choose a current model ID
```
curl https://api.gonkagate.com/v1/models \
  -H "Authorization: Bearer gp-your-api-key"
```
**JSON Example**
```
{
  "object": "list",
  "data": [
    {
      "id": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
      "object": "model",
      "created": 0,
      "owned_by": "gonka"
    }
  ]
}
```

Use one `data[].id` value as `model` in your next request.

## Use the ID in a request
**Request Example**
```
curl https://api.gonkagate.com/v1/chat/completions \
  -H "Authorization: Bearer gp-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    "messages": [
      { "role": "user", "content": "Summarize this document." }
    ]
  }'
```

Use any current `data[].id` from `GET /v1/models`. Keep `model` configurable so you can refresh it without changing application code.

## Before you set a default model

- An API key that can call `GET /v1/models`
- One or two real prompts from your workload
- A config or environment variable for the default `model` value
- A note on whether the workload depends on tools, vision, structured outputs, or streaming

## Common model ID issues

- `404 model_not_found` means the ID is no longer valid for the request you sent. Refresh it from `GET /v1/models`.
- `GET /v1/models` can temporarily return `503` while the catalog is not ready. Retry with bounded backoff. Do not scrape `/models` or guess IDs manually.
- `:online` is not a canonical model ID returned by `GET /v1/models`. Start with the base `data[].id`. Append `:online` only as a request-time shorthand for web search on that call. For new GonkaGate integrations, prefer `plugins: [{ "id": "web" }]`.
- `/models` is the human-readable catalog. It helps you browse, but `GET /v1/models` is the machine-readable source for application code.
- `GET /v1/models` returns minimal OpenAI-compatible model objects. Use `/pricing` for billing rules and `/models` for catalog context such as display names and capability notes.
- If your workload depends on tools, vision, or structured outputs, validate that capability with the model you plan to ship before rollout.
- A default model is not a permanent contract. Recheck it before releases, migrations, or fallback changes.

## See also

- [Quickstart: first request](/en/docs/quickstart) to send a working request with the model ID you selected here.
- [GET /v1/models — List Available Models](/en/docs/api/api-reference/models/get-models) for the exact response schema.
- [Live Models Catalog](/en/models) to browse current models before you copy an ID.
- [Pricing](/en/pricing) if model choice is blocked by billing rules rather than API compatibility.