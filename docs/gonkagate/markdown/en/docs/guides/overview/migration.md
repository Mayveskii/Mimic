# OpenAI to GonkaGate Migration

Move an existing OpenAI-compatible chat.completions integration to GonkaGate.

Move an existing OpenAI-compatible `chat.completions` integration to GonkaGate by changing the base URL, API key, and model ID, then verify the switch with `GET /v1/models` and one `POST /v1/chat/completions` request. Keep the migration narrow: change transport first, prove parity, then touch prompts or broader app logic.

## Before you switch

- A verified GonkaGate account
- A saved `gp-...` API key
- A fresh model ID from `GET /v1/models` for the workload you are migrating
- Enough prepaid USD balance for the smoke test and rollout
- A working OpenAI-compatible path in your app

## Change only three values

- Set `base_url` or `baseURL` to `https://api.gonkagate.com/v1`
- Replace the OpenAI key in `sk-...` format with a GonkaGate key in `gp-...` format
- Replace the OpenAI model name with a fresh GonkaGate model ID from `GET /v1/models`

**Request Example**
```
 from openai import OpenAI

client = OpenAI(
-    base_url="https://api.openai.com/v1",
+    base_url="https://api.gonkagate.com/v1",
-    api_key="sk-..."
+    api_key="gp-..."
)

response = client.chat.completions.create(
-    model="gpt-5.2",
+    model="qwen/qwen3-235b-a22b-instruct-2507-fp8",
     messages=[{"role": "user", "content": "Say hello from GonkaGate"}]
)
```

## Verify the migration
 Verify the migration
```
curl https://api.gonkagate.com/v1/models \
  -H "Authorization: Bearer gp-your-api-key"
```
 Command
```
curl https://api.gonkagate.com/v1/chat/completions \
  -H "Authorization: Bearer gp-your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    "messages": [
      { "role": "user", "content": "Say hello from GonkaGate" }
    ]
  }'
```

Expected result:

- `GET /v1/models` returns a model list
- `POST /v1/chat/completions` returns a normal response with `choices[0].message.content`

## Common migration failures

| Response | What it usually means | What to do |
|---|---|---|
| 401 invalid_api_key | The Bearer header, key value, or account state is wrong | Recheck the Authorization: Bearer gp-... header and confirm the key is active |
| 404 model_not_found | The model ID is stale or unsupported | Refresh the model ID from GET /v1/models before treating the migration as broken |
| 429 insufficient_quota | The account does not have enough available balance | Add balance or lower test traffic before rollout |
| 429 rate_limit_exceeded | You hit a runtime limit | Handle it with bounded retries instead of treating it as a migration-only failure |
| 5xx | A temporary upstream or platform error occurred | Retry with bounded backoff and confirm whether the failure persists |

## Compatibility notes

- Keep strict response typing aligned with the OpenAI chat-completions shape. GonkaGate-specific billing totals are not returned in each API response; use dashboard activity for completed usage and spend.
- This guide covers `chat.completions` and `GET /v1/models`. If your app depends on embeddings, Responses API, Assistants, Audio, Batch, or fine-tuning, keep those flows on a different provider path for now.
- Streaming, tools, JSON mode, and vision depend on the selected model and API surface. Confirm support for the exact model before rollout.

## See also

- [OpenAI SDK Compatibility for GonkaGate](/en/docs/sdk/openai) to move this switch into production application code
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, storage, rotation, and auth-specific failures
- [Model Selection Guide](/en/docs/guides/overview/models) for choosing and refreshing GonkaGate model IDs
- [API Reference Overview](/en/docs/api/reference/overview) for request fields, streaming behavior, and retry policy after the migration works
- [Pricing](/en/pricing) for prepaid USD billing rules and cost breakdown