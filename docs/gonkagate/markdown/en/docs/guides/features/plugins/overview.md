# Chat Completions Plugins

Choose the right plugin for a /v1/chat/completions request.

Use `plugins` in `/v1/chat/completions` when one request needs web grounding, an explicit PDF engine override, server-side response healing for structured output, or secret sanitization before model execution.

## Add a plugin to the request

Add one entry to `plugins`. Keep the rest of the request as a normal chat-completions call.
 request.json
```
{
  "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  "messages": [
    {
      "role": "user",
      "content": "Summarize this incident note, but hide apiKey=sk-live-demo-123456."
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

- `web`: add `{ "id": "web" }`. This plugin is explicit opt-in only.
- `file-parser`: add `pdf.engine` only when the request must force `pdf-text` or `native`.
- `response-healing`: also send `stream: false` and a structured `response_format`.
- `privacy-sanitization`: also send `mode` as `redact`, `tokenize`, or `block`.

## Available plugins

| Plugin | Description | Docs |
|---|---|---|
| Web Search | Built-in web grounding for one request. Use `plugins: [{ id: "web" }]` or the OpenRouter-compatible `:online` model suffix. | Web Search |
| PDF Inputs | Force a specific PDF engine when server-managed auto handling is not enough, or keep reading for `pdf-text` versus `native`. | PDF Inputs |
| Response Healing | Add a server-side repair step for non-stream structured output when `response_format` is already correct but responses still break. | Response Healing |
| Privacy Sanitization | Redact, tokenize, or block sensitive fragments before model execution. Do not combine it with built-in web search. | Privacy Sanitization |

## How activation works

- `plugins` in the request is the normal activation path for `/v1/chat/completions`.
- Authenticated API requests may also inherit saved plugin defaults when plugin settings policy is enabled.
- `web` still requires explicit opt-in per request.
- `file-parser` can still be applied automatically for PDF inputs when server-managed engine selection is enough.

## Limits and compatibility

- `web` and `privacy-sanitization` cannot run in the same request.
- `response-healing` only works with `stream: false` and `response_format.type` set to `json_object` or `json_schema`.
- `web` supports streaming and non-streaming requests.
- `privacy-sanitization` supports streaming and non-streaming requests.

## See also

- [Web Search](./web-search) for explicit built-in web grounding.
- [PDF Inputs](./pdf-inputs) for `file-parser`, `pdf-text` vs `native`, and PDF-specific limits.
- [Response Healing](./response-healing) for non-stream structured-output repair.
- [Privacy Sanitization](./privacy-sanitization) for request-level secret sanitization.
- [Structured Outputs](/en/docs/guides/features/structured-outputs) if the real job is the JSON contract rather than plugin selection.
- [Tool Calling](/en/docs/guides/features/tool-calling) if your application should own retrieval or tool execution.
- [Chat Completions API reference](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact `plugins` field schema.