# API Reference Overview

Base URL, Bearer auth, core endpoints, OpenAI-compatible request format, and response basics for the GonkaGate API.

Use the GonkaGate API at `https://api.gonkagate.com/v1` with `Authorization: Bearer gp-...`, send generation requests to `POST /v1/chat/completions`, and refresh model IDs from `GET /v1/models`. Requests use the OpenAI-compatible chat format and return either JSON or SSE. Need the first working request? Start with [Quickstart](/en/docs/quickstart).

## API basics

- Base URL: `https://api.gonkagate.com/v1`
- Auth: `Authorization: Bearer gp-...`
- Main generation endpoint: `POST /v1/chat/completions`
- Model list endpoint: `GET /v1/models`
- Non-streaming responses: JSON with `choices` and `usage`
- Streaming responses: add `stream: true` and parse Server-Sent Events (SSE)

## Request skeleton

Use this shape for any basic `POST /v1/chat/completions` call. Replace `model` with a fresh ID from `GET /v1/models`.
**Request Example**
```
export GONKAGATE_API_KEY="gp-your-api-key"

curl https://api.gonkagate.com/v1/chat/completions \
  -H "Authorization: Bearer $GONKAGATE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "<model-id-from-get-v1-models>",
    "messages": [
      {
        "role": "user",
        "content": "Reply with the word pong."
      }
    ]
  }'
```

The minimum request is base URL, Bearer auth, a current model ID, and OpenAI-compatible `messages`.

## Request rules

- Send JSON to `POST /v1/chat/completions`.
- Include `Authorization: Bearer gp-...` and `Content-Type: application/json`.
- Set `model` to a fresh ID from `GET /v1/models`.
- Send `messages` in the OpenAI-compatible chat format.
- Add `stream: true` only when your client already parses SSE.

## Response and failure basics

- Non-streaming responses are JSON. Read the model output from `choices[0].message.content`.
- Streaming responses arrive as SSE `data:` events and end with `[DONE]`.
- Classify failures by `HTTP status + error.code`, not status alone.
- Keep `x-request-id` when a request fails or behaves unexpectedly.

## Common mistakes

- Using a stale model ID instead of refreshing it from `GET /v1/models`.
- Debugging auth, model selection, and request body shape at the same time. Check them in that order.
- Treating every `429` as retryable throttling. `insufficient_quota` is a billing state, not a backoff case.
- Treating this overview as the full schema. Exact request and response fields live on the endpoint reference pages.

## See also

- [Create a chat completion](/en/docs/api/api-reference/chat/send-chat-completion-request) for the full `POST /v1/chat/completions` request schema, JSON vs SSE responses, error payloads, and `x-idempotency-key`.
- [GET /v1/models](/en/docs/api/api-reference/models/get-models) for current machine-readable model IDs.
- [Streaming Responses](/en/docs/api/reference/streaming) for `stream: true`, SSE parsing, and end-of-stream handling.
- [API Error Handling](/en/docs/api/reference/error-handling) for retry, stop, and escalation rules.
- [Authentication and API Keys](/en/docs/authentication/api-keys) if the Bearer header or key state is still in doubt.