# OpenAI SDK Compatibility

Switch the base URL, API key, and model ID to use the official OpenAI SDK with GonkaGate.

Change only the base URL, API key, and model ID to use the official OpenAI SDK with GonkaGate. Keep the same client shape, `messages`, and most of the same app logic for `POST /v1/chat/completions` and `GET /v1/models`.

## Minimum working change

Change the three connection values GonkaGate owns: base URL, API key, and model ID.
**Endpoint**
```
 client = OpenAI(
-    base_url / baseURL = "https://api.openai.com/v1",
+    base_url / baseURL = "https://api.gonkagate.com/v1",
-    api_key = "sk-..."
+    api_key = "gp-..."
 )

 response = client.chat.completions.create(
-    model = "gpt-..."
+    model = "<current-gonkagate-model-id>"
 )
```

Expected result: your existing `chat.completions.create(...)` call still works after you swap the connection values and use a current GonkaGate model ID.

Keep strict response typing aligned with the OpenAI chat-completions shape. GonkaGate-specific billing totals are not returned in each API response; use the dashboard for completed usage and spend.

## What you need before you run it

- A verified GonkaGate account
- A saved `gp-...` API key in your server-side secret store
- A current GonkaGate model ID instead of an OpenAI default model name
- Enough prepaid USD balance for the request

Use [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, storage, and rotation.

## Common errors and current limits

- `401 invalid_api_key` usually means the Bearer value, key state, or account state is wrong.
- `404 model_not_found` means the model ID is stale or not supported on GonkaGate.
- `429 insufficient_quota` means the available prepaid USD balance is too low for the request.
- `429 rate_limit_exceeded` and `5xx` should be handled as runtime failures with bounded retries.
- This path covers the official OpenAI SDK for `chat.completions` plus `GET /v1/models`.
- If your app depends on embeddings, Responses API, Assistants, Audio, Batch, or fine-tuning, keep those flows on a different provider path for now.
- Streaming, tools, JSON mode, and vision depend on the selected model and the API surface you use.

## See also

- [SDK Guides for GonkaGate](/en/docs/sdk) to choose the runtime-specific guide for Python, TypeScript, Go, .NET, or Java.
- [Model Selection Guide](/en/docs/guides/overview/models) to choose and refresh GonkaGate model IDs.
- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you already have working OpenAI code and need rollout-oriented switch steps.
- [Framework & Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview) if you use LangChain, Vercel AI SDK, or another wrapper instead of the raw OpenAI client.
- [Claude Code](/en/docs/guides/coding-agents/claude-code), [Cursor](/en/docs/guides/coding-agents/cursor), [Kilo Code](/en/docs/guides/coding-agents/kilo-code), [MiMoCode](/en/docs/guides/coding-agents/mimo-code), [OpenCode](/en/docs/guides/coding-agents/opencode), [OpenClaw](/en/docs/guides/coding-agents/openclaw), and [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) if you use an agent-owned tool path.
- [API Reference Overview](/en/docs/api/reference/overview) for exact request fields, streaming behavior, and runtime-independent failure policy.