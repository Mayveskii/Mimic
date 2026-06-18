# Web Search

Enable explicit web search grounding for a chat-completions request.

Enable web search per request with `plugins: [{ "id": "web" }]`. The `model: "...:online"` suffix is also accepted for OpenRouter compatibility, but new GonkaGate integrations should prefer the explicit plugin object.

## Minimal request
 request.json
```
{
  "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  "messages": [
    {
      "role": "user",
      "content": "What changed in the latest Gonka Network release?"
    }
  ],
  "plugins": [
    {
      "id": "web"
    }
  ]
}
```

When `engine` is omitted, gonka-proxy uses its configured internal web-search provider.

## Optional fields
 request.json
```
{
  "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  "messages": [
    {
      "role": "user",
      "content": "Find recent sources and summarize the facts."
    }
  ],
  "plugins": [
    {
      "id": "web",
      "max_results": 5,
      "search_prompt": "Prefer primary sources and recent documentation."
    }
  ]
}
```

- `max_results`: requested result count. The server may cap it to a safe maximum.
- `search_prompt`: grounding prompt override used when formatting retrieved web results for the model.
- `engine`: OpenRouter-compatible engine hint. Unsupported values are rejected.
- `enabled: false`: disables web search when a web plugin object is present.

## Compatibility

- Web search works for streaming and non-streaming chat completions.
- `web` is explicit opt-in only. Authenticated saved plugin defaults do not enable it automatically.
- Do not combine `web` with `privacy-sanitization`; that request fails with HTTP 400 and code `web_search_privacy_sanitization_not_supported`.
- `web_search_options` is accepted for OpenRouter migration compatibility and is currently ignored in v1.

## `:online` shorthand

This request is equivalent explicit opt-in:
 request.json
```
{
  "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8:online",
  "messages": [{ "role": "user", "content": "Find recent sources." }]
}
```

Use the base model ID from `GET /v1/models`; append `:online` only at request time when you intentionally want web grounding.

## Related pages

- [Chat Completions Plugins](./overview) for plugin selection and compatibility.
- [Privacy Sanitization Plugin](./privacy-sanitization) for the incompatible secret-sanitization path.
- [Chat Completions API reference](/en/docs/api/api-reference/chat/send-chat-completion-request) for the full request schema.