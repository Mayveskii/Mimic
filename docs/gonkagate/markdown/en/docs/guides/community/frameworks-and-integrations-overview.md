# Framework and Tool Guides

Choose the GonkaGate guide for your framework or coding tool.

Choose the guide for the wrapper or community tool your app already uses for OpenAI-compatible calls. If you need Claude Code, Cursor, Kilo Code, MiMoCode, OpenCode, OpenClaw, or Hermes Agent, use the dedicated guides for [Claude Code](/en/docs/guides/coding-agents/claude-code), [Cursor](/en/docs/guides/coding-agents/cursor), [Kilo Code](/en/docs/guides/coding-agents/kilo-code), [MiMoCode](/en/docs/guides/coding-agents/mimo-code), [OpenCode](/en/docs/guides/coding-agents/opencode), [OpenClaw](/en/docs/guides/coding-agents/openclaw), or [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) instead.

If you are wiring the raw OpenAI SDK directly, use [OpenAI SDK Compatibility](/en/docs/sdk/openai) instead.

## Shared OpenAI-compatible values
 Shared OpenAI-compatible values
```
api_key = gp-your-api-key
base_url = https://api.gonkagate.com/v1
model = current GonkaGate model ID
```

These values apply to LangChain, LlamaIndex, PydanticAI, TanStack AI, Vercel AI SDK, Aider, Cline, and Roo Code. The `n8n` community node uses the same GonkaGate API surface, but the package setup is documented separately below.

Use a fresh model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models) before you send real traffic on the OpenAI-compatible path.

## Framework guides

| Guide | Best for | Docs |
|---|---|---|
| LangChain | Apps already using `ChatOpenAI` in Python or TypeScript. Keeps the wrapper and only swaps the connection values. | LangChain guide |
| LlamaIndex | Apps already using the OpenAI LLM wrapper or `Settings.llm`. Focuses on the narrow `OpenAI(...)` configuration path. | LlamaIndex guide |
| PydanticAI | Apps already using `OpenAIProvider` and `OpenAIChatModel`. Covers the provider and model swap, not broader agent design. | PydanticAI guide |
| TanStack AI | Apps already following TanStack chat patterns. Covers the OpenAI adapter plus the server-route wiring TanStack needs. | TanStack AI guide |
| Vercel AI SDK | Apps already using AI SDK primitives such as `streamText`. Covers the OpenAI-compatible provider setup, not the raw `openai` client. | Vercel AI SDK guide |

## Community tool guides

| Guide | Best for | Docs |
|---|---|---|
| Aider | Local repo chat sessions through GonkaGate. Covers a one-session CLI command or a repo-local config file. | Aider guide |
| Cline | GonkaGate inside the VS Code assistant flow. Covers provider setup plus a first validation prompt. | Cline guide |
| Roo Code | Roo's OpenAI-compatible provider calling GonkaGate. Covers the provider switch plus a read-only validation step. | Roo Code guide |

## Community integrations

| Guide | Best for | Docs |
|---|---|---|
| n8n | Self-hosted n8n that should use the official GonkaGate community node package instead of piecing together generic OpenAI-compatible nodes by hand. | n8n guide |

## Before you change the rest of your app

- Verify one small request or status check first.
- If requests still fail, confirm that the wrapper or tool is still using its OpenAI-compatible path, or the curated Anthropic-compatible Claude Code path.
- These guides cover the wrapper or tool layer. For the raw SDK path, full migration steps, or exact request fields, use the pages below.

## See also

- [GonkaGate Quickstart](/en/docs/quickstart) if you still need your first successful request from scratch.
- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you are switching an existing OpenAI-compatible app end to end.
- [API Reference Overview](/en/docs/api/reference/overview) for exact request fields, streaming behavior, and runtime-independent failure policy.
- [Claude Code](/en/docs/guides/coding-agents/claude-code), [Cursor](/en/docs/guides/coding-agents/cursor), [Kilo Code](/en/docs/guides/coding-agents/kilo-code), [MiMoCode](/en/docs/guides/coding-agents/mimo-code), [OpenCode](/en/docs/guides/coding-agents/opencode), [OpenClaw](/en/docs/guides/coding-agents/openclaw), and [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) for current agent-owned setup paths.