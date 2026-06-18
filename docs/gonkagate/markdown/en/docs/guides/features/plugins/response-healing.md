# Response Healing

Use response-healing to repair malformed non-stream structured output before it reaches your client.

`response-healing` adds a server-side repair step for malformed non-stream structured output. Use it after your base `json_object` or `json_schema` request already works and you want fewer JSON parse failures before the response reaches your client.

## Minimum working request
 request.json
```
{
  "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  "messages": [
    { "role": "system", "content": "Return valid JSON only." },
    { "role": "user", "content": "Extract { title, priority } from: \"Fix login bug (high)\"." }
  ],
  "stream": false,
  "response_format": { "type": "json_object" },
  "plugins": [{ "id": "response-healing" }]
}
```

**Expected result:** GonkaGate tries to repair malformed structured output before returning the final message. Your client still reads and parses `choices[*].message.content`.

## Use it when

- `stream` is `false`.
- `response_format.type` is `json_object` or `json_schema`.
- Your base structured-output request already works, but you want extra server-side hardening before client parsing or schema validation.

## Do not use it when

- The response is streaming.
- The job is free-form text rather than structured output.
- The task depends on tool calls or other non-string content. `response-healing` does not repair those outputs.

## Turn it on or off

Request-level opt-in works when all of these are true:

- `stream` is `false`.
- `response_format.type` is `json_object` or `json_schema`.
- `plugins` includes `{ "id": "response-healing" }`.
- The plugin entry is not disabled with `enabled: false`.

To disable it for one request, remove the plugin entry or keep it and send `enabled: false`.

For OpenRouter-compatible clients, only `id` and optional `enabled` affect behavior in GonkaGate today. Compatibility extras such as `mode` or `schema_validation` are accepted, but ignored for healing activation.

## Saved plugin settings and overrides

Authenticated requests can also inherit saved plugin settings when plugin settings policy is enabled.

- `plugins: [{ "id": "response-healing" }]` is the explicit request-level contract, but not the only activation source.
- Saved defaults are inherited only when the request-level plugin entry is absent.
- `enabled: false` disables healing for this request only when the saved setting is not locked.
- If a locked saved setting conflicts with the request, GonkaGate returns `400 invalid_request_error` with `code=plugin_override_blocked`.

## SDK examples

### TypeScript

Official `openai` TypeScript types do not expose GonkaGate-specific `plugins`. Use the documented workaround for undocumented request params.
 response-healing.ts
```
import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "https://api.gonkagate.com/v1",
  apiKey: "gp-your-api-key"
});

const response = await client.chat.completions.create({
  model: "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  stream: false,
  messages: [
    { role: "system", content: "Return JSON only." },
    { role: "user", content: "Extract title and priority." }
  ],
  response_format: { type: "json_object" },
  // @ts-expect-error undocumented request field for an OpenAI-compatible backend
  plugins: [{ id: "response-healing" }]
});

const content = response.choices[0]?.message?.content ?? "{}";
const data = JSON.parse(content);
console.log(data);
```

### Python

Pass `plugins` through `extra_body`.
 response-healing.py
```
from openai import OpenAI
import json

client = OpenAI(
    base_url="https://api.gonkagate.com/v1",
    api_key="gp-your-api-key"
)

response = client.chat.completions.create(
    model="qwen/qwen3-235b-a22b-instruct-2507-fp8",
    stream=False,
    messages=[
        {"role": "system", "content": "Return JSON only."},
        {"role": "user", "content": "Extract title and priority."}
    ],
    response_format={"type": "json_object"},
    extra_body={"plugins": [{"id": "response-healing"}]}
)

parsed = json.loads(response.choices[0].message.content)
print(parsed)
```

## Before rollout

- You still parse JSON on the client.
- Log `x-request-id` for support and debugging.
- If you retry non-stream requests, use `x-idempotency-key` on the active healing path.

## Errors to handle

| Error | What it means | What to do next |
|---|---|---|
| 502 response_healing_failed | The model returned malformed structured output and the repair path still failed. | Inspect the prompt or schema, retry with a bounded policy, or test another model/provider pair. |
| 502 response_schema_validation_failed | Runtime could not validate the final output against the json_schema: either the output still does not satisfy the schema, or the schema payload itself is not usable for validation or compilation. | Check the schema payload first, then simplify the schema, relax strictness where appropriate, or fall back to json_object plus client validation. |
| 503 service_unavailable | Under overload, direct parse failed and the heavier healing path was skipped before repair could run. | Retry with bounded backoff and use x-idempotency-key on non-stream retries. |
| 400 plugin_override_blocked | Locked saved plugin settings conflict with the current request override. | Remove the incompatible override or align the request with the saved plugin policy. |
For invalid payload types such as `enabled: "yes"`, expect the normal `400 invalid_request_error` schema-validation failure with `code=invalid_request`.

## See also

- [Structured Outputs](/en/docs/guides/features/structured-outputs) for choosing between `json_object` and `json_schema`.
- [Plugins Overview](/en/docs/guides/features/plugins/overview) for choosing between chat-completions plugins.
- [Chat Completions API reference](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact request schema.