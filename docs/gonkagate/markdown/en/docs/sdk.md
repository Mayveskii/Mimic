# OpenAI SDK Guides

Choose the runtime guide for the official OpenAI SDK with GonkaGate.

Choose the runtime guide for the official OpenAI SDK after you have one working GonkaGate request. If you use LangChain, Vercel AI SDK, Aider, Cline, Roo Code, or another wrapper or community tool, use [Framework & Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview) instead. If you use Claude Code, Cursor, Kilo Code, MiMoCode, OpenCode, OpenClaw, or Hermes Agent, use the dedicated [Claude Code](/en/docs/guides/coding-agents/claude-code), [Cursor](/en/docs/guides/coding-agents/cursor), [Kilo Code](/en/docs/guides/coding-agents/kilo-code), [MiMoCode](/en/docs/guides/coding-agents/mimo-code), [OpenCode](/en/docs/guides/coding-agents/opencode), [OpenClaw](/en/docs/guides/coding-agents/openclaw), or [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) page instead.

## Shared OpenAI SDK shape
**Request Example**
```
client = OpenAI-compatible SDK client
base_url / baseURL = "https://api.gonkagate.com/v1"
api_key = "gp-your-api-key"

response = client.chat.completions.create(
  model = "current GonkaGate model ID",
  messages = [...]
)
```

Expected result: keep the same `chat.completions.create(...)` request shape in Python, TypeScript, Go, .NET, or Java. Only client setup and runtime-specific ergonomics change.

## Before you open a runtime guide

- One successful GonkaGate request or migration smoke test
- A saved `gp-...` API key that your app can read from a secret store or environment variable
- A current GonkaGate model ID for the workload you are integrating
- A server-side place to keep the key out of browsers and public bundles

## Choose a runtime guide

- [Python SDK](/en/docs/sdk/python): backend services, jobs, and scripts with sync or async clients.
- [TypeScript SDK](/en/docs/sdk/typescript): Node.js, Deno, Bun, and edge-friendly runtimes.
- [Go SDK](/en/docs/sdk/go): `context.Context`, explicit timeouts, and service-level client reuse.
- [.NET SDK](/en/docs/sdk/dotnet): Betalgo.OpenAI with DI-friendly client setup.
- [Java SDK](/en/docs/sdk/java): `openai-java` with explicit client configuration.

## Common SDK errors and compatibility notes

- `401 invalid_api_key` usually means the key is missing, malformed, or unavailable in the runtime where the SDK runs.
- `404 model_not_found` usually means the SDK setup is fine but the model ID is stale.
- `429 insufficient_quota` means the prepaid USD balance is too low for the request.
- Keep response handling aligned with the OpenAI chat-completions shape. Use dashboard activity for completed usage and spend.

## See also

- [OpenAI SDK Compatibility for GonkaGate](/en/docs/sdk/openai) for the shared base URL, API key, model ID, and response-handling rules across runtimes.
- [Authentication and API Keys](/en/docs/authentication/api-keys) to create, store, and rotate `gp-...` keys.
- [GonkaGate Quickstart](/en/docs/quickstart) if you still need one verified request before SDK wiring.
- [Migrate from OpenAI to GonkaGate](/en/docs/guides/overview/migration) if you are switching an already-working OpenAI app.
- [Framework & Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview) if you integrate through LangChain, Vercel AI SDK, Aider, Cline, Roo Code, or another wrapper or community tool.
- [Claude Code](/en/docs/guides/coding-agents/claude-code), [Cursor](/en/docs/guides/coding-agents/cursor), [Kilo Code](/en/docs/guides/coding-agents/kilo-code), [MiMoCode](/en/docs/guides/coding-agents/mimo-code), [OpenCode](/en/docs/guides/coding-agents/opencode), [OpenClaw](/en/docs/guides/coding-agents/openclaw), and [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) if you use an agent-owned tool path.