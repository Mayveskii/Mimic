# Privacy Sanitization

Redact, tokenize, or block secrets with privacy-sanitization.

Use `privacy-sanitization` to sanitize secrets before `/v1/chat/completions` reaches the model. Use `redact` for readable masked text, `tokenize` for stable per-request placeholders, and `block` to stop the request before execution. For broader security and data-handling policy, see [Security](/en/security).

## Send the minimum request
 request.json
```
{
  "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  "messages": [
    {
      "role": "user",
      "content": "Summarize this incident note, but hide apiKey=sk-live-demo-123456 and DATABASE_URL=postgres://app:super-secret@db.internal:5432/gonkagate."
    }
  ],
  "plugins": [
    {
      "id": "privacy-sanitization",
      "mode": "redact"
    }
  ]
}
```

Expected result: the request continues with sanitized input, and covered fragments are not forwarded to the model in raw form.

## Choose the mode

| Mode | What changes before execution | What the model receives | Best for | Avoid when |
|---|---|---|---|---|
| redact | Covered fragments are replaced with generic masks or redaction labels. | Sanitized text with readable placeholders. | Public chat, support prompts, or operator-facing flows where the prompt should stay readable after sanitization. | The model must distinguish repeated hidden values inside one request. |
| tokenize | Covered fragments are replaced with stable placeholders. | Sanitized text with stable tokens that preserve which hidden value is which inside that request. | Single requests where the same hidden value appears multiple times and the model must keep those references straight. | Human-facing prompts that should read naturally, or flows that expect server-managed history or session restore later. |
| block | The request is rejected before model execution. | Nothing. Model execution does not start. | Strict compliance, no-risk flows, or cases where continuing with sanitized content would still be unsafe or misleading. | UX flows where users expect automatic sanitization and continuation. |

## Compare one input across all three modes

Placeholder strings below are examples only. Exact replacement strings are implementation-defined.

**Input**
 Compare one input across all three modes
```
Compare sk-live-prod-123 and sk-live-staging-456. In the next step, use sk-live-prod-123 again.
```

**`redact`**
 redact
```
Compare [SECRET:API_KEY] and [SECRET:API_KEY]. In the next step, use [SECRET:API_KEY] again.
```

**`tokenize`**
 tokenize
```
Compare [TOKEN_1] and [TOKEN_2]. In the next step, use [TOKEN_1] again.
```

**`block`**
 block
```
Request rejected before model execution.
```

- Choose `redact` when the prompt should stay readable after sanitization.
- Choose `tokenize` when one request needs stable hidden placeholders for repeated secret references.
- Choose `block` when continuing with sanitized content would still be unsafe.

`redact` keeps the request readable but collapses secret identity. `tokenize` keeps the value hidden while preserving which secret is which inside the same sanitized request.

Use `tokenize` to preserve hidden-value identity inside one sanitized request, not to format or restore chat history.

## Know the limits before rollout

- `mode` is required.
- Public `privacy-sanitization` on default `/v1/chat/completions` is stateless: sanitization runs on request ingress and does not keep privacy state across requests.
- `redact`, `tokenize`, and `block` work in both streaming and non-streaming requests.
- Do not expect `_gonka.privacy.history_writeback`, `event: gonka.input_patch`, raw restore in `tool_calls`, or any other client-visible history restore path.
- Do not rely on history or session state from this default public surface.
- This public managed profile is secret-oriented. It does not promise broad generic PII redaction by default.
- `privacy-sanitization` is currently incompatible with built-in web search, including `plugins: [{ "id": "web" }]` and `model: "...:online"`. Choose one path per request.
- Authenticated requests can also inherit saved plugin settings when plugin settings policy is enabled.

## Enable or disable the plugin

Request-level opt-in works when:

- `plugins` includes `id="privacy-sanitization"`
- `mode` is `redact`, `tokenize`, or `block`
- that plugin entry has `enabled !== false`

 plugins.json
```
{
  "plugins": [
    {
      "id": "privacy-sanitization",
      "mode": "redact"
    }
  ]
}
```

Authenticated requests can also activate `privacy-sanitization` from saved plugin settings. Treat the request payload as the request-level contract, but not the only activation source when plugin settings policy is enabled.

To disable request-level activation for one request:

1. Omit `privacy-sanitization` from `plugins` if you do not use saved plugin defaults.
2. If saved plugin settings may auto-activate it, send `enabled: false` or disable the saved default.
3. If saved plugin settings are locked by policy, request-level overrides may be rejected.

## Use it from SDKs

The OpenAI SDK request types do not model GonkaGate-specific `plugins`. Send the field through the runtime-specific escape hatch.

### TypeScript

Use `// @ts-expect-error` or your own wrapper around undocumented request fields.
 privacy-sanitization.ts
```
import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "https://api.gonkagate.com/v1",
  apiKey: "gp-your-api-key"
});

const response = await client.chat.completions.create({
  model: "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  messages: [
    {
      role: "user",
      content:
        "Summarize this incident note, but hide Authorization: Bearer demo-secret-token and DATABASE_URL=postgres://app:super-secret@db.internal:5432/gonkagate."
    }
  ],
  // @ts-expect-error undocumented request field for an OpenAI-compatible backend
  plugins: [{ id: "privacy-sanitization", mode: "redact" }]
});

console.log(response.choices[0]?.message?.content);
```

### Python

Use `extra_body` for undocumented request fields.
 privacy-sanitization.py
```
from openai import OpenAI

client = OpenAI(
    base_url="https://api.gonkagate.com/v1",
    api_key="gp-your-api-key"
)

response = client.chat.completions.create(
    model="qwen/qwen3-235b-a22b-instruct-2507-fp8",
    messages=[
        {
            "role": "user",
            "content": "Summarize this incident note, but hide apiKey=sk-live-demo-123456."
        }
    ],
    extra_body={
        "plugins": [{"id": "privacy-sanitization", "mode": "redact"}]
    }
)

print(response.choices[0].message.content)
```

## Handle common failures

| Condition | Result | What to do |
|---|---|---|
| privacy-sanitization is combined with built-in web search (plugins: [{ id: "web" }] or model: "...:online") | Usually HTTP 400 invalid_request_error with code web_search_privacy_sanitization_not_supported. If web search is unavailable, the request can return HTTP 503 server_error with code web_search_unavailable instead. | Choose one path per request: built-in web grounding or privacy sanitization. |
| The privacy sidecar is unavailable | HTTP 503 privacy_sanitization_unavailable; the raw prompt is not forwarded upstream | Treat this as fail-closed behavior. Do not assume passthrough. |
| mode: "block" detects a covered fragment | The request is rejected before model execution | Surface it as an expected control path and let the caller fix the input or choose another mode. |

## See also

- [Plugins Overview](./overview) to compare plugin behavior and saved defaults.
- [Web Search](./web-search) for the separate built-in grounding path. It is currently incompatible with `privacy-sanitization`; choose one path per request.
- [Chat Completions API reference](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact plugin field schema.
- [Security](/en/security) for broader security and data-handling policy.