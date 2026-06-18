# Quickstart: first request

GonkaGate developer docs for sending your first OpenAI-compatible request.

Send one OpenAI-compatible `POST /v1/chat/completions` request through GonkaGate to verify your access to Gonka Network. You need `https://api.gonkagate.com/v1`, `Authorization: Bearer gp-...`, and a fresh model ID from [`GET /v1/models`](/en/docs/api/api-reference/models/get-models).

## Before you start

- A verified GonkaGate account
- A `gp-...` API key; use [Authentication and API Keys](/en/docs/authentication/api-keys) if you need to create one
- Save the key when it is shown. The full secret is not revealed again later.
- Enough available balance for the request

Need secure storage or rotation? Use [Authentication and API Keys](/en/docs/authentication/api-keys).

## Send the request

Use a fresh model ID from [`GET /v1/models`](/en/docs/api/api-reference/models/get-models). If the endpoint returns `503`, retry with short bounded backoff. Do not copy IDs from `/models` HTML and do not guess them manually.
**Endpoint**
```
export GONKAGATE_API_KEY="gp-your-api-key"

curl https://api.gonkagate.com/v1/chat/completions \
  -H "Authorization: Bearer $GONKAGATE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    "messages": [
      { "role": "user", "content": "Hello, Gonka!" }
    ]
  }'
```

## Expected result

A successful request returns `200 OK`. The model output is in `choices[0].message.content`.

## Common first failures

| Response | What it usually means | What to check next |
|---|---|---|
| 401 invalid_api_key | The auth header is missing or malformed, the key is invalid, disabled, or expired, or the account email is still unverified. | Authentication and API Keys |
| 404 model_not_found | The model ID is stale or not currently available. | Model Selection Guide |
| 429 insufficient_quota | The account does not have enough available balance for the request. | Pricing |

## See also

- [OpenAI SDK for GonkaGate](/en/docs/sdk/openai) to move this request into application code.
- [API Reference Overview](/en/docs/api/reference/overview) for exact request fields, streaming, retries, and runtime behavior after the first call works.
- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you already have a working OpenAI-compatible integration and only need the GonkaGate switch.