# TypeScript SDK

Send chat.completions requests to GonkaGate from TypeScript with the official OpenAI SDK.

Send one `chat.completions` request from TypeScript with the official OpenAI SDK, then adapt the same client shape to your runtime.

## Minimum working example

Install `openai`, set `GONKAGATE_API_KEY`, and run one request.
 npmpnpmyarnbunCommand
```
npm install openai
```
**Request Example**
```
import OpenAI from "openai";

const apiKey = process.env.GONKAGATE_API_KEY;

if (!apiKey) {
  throw new Error("Set GONKAGATE_API_KEY before running this example.");
}

const client = new OpenAI({
  baseURL: "https://api.gonkagate.com/v1",
  apiKey
});

async function main() {
  const response = await client.chat.completions.create({
    model: "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    messages: [{ role: "user", content: "Hello, GonkaGate!" }]
  });

  console.log(response.choices[0]?.message?.content);
}

main().catch(console.error);
```

Expected result: you get text in `response.choices[0]?.message?.content`.

Use a current GonkaGate model ID before you send real traffic. The example value above is only illustrative.

## What you need before you use it

- `GONKAGATE_API_KEY` available in the runtime where this script runs
- A current GonkaGate model ID for the workload you want to call
- A server-side place to keep the key out of browsers and public bundles

## TypeScript-specific notes

- The official SDK supports Node.js, Deno, Bun, Cloudflare Workers, and Vercel Edge Runtime.
- The example uses `process.env` for the Node.js or Bun happy path. In Deno or edge runtimes, keep the same `baseURL` and request shape, but read the key from that platform’s secret API.
- Browser use is disabled by default. Do not enable `dangerouslyAllowBrowser` for public client-side apps with a real `gp-...` key.
- Set `timeout` and `maxRetries` explicitly on the client before production traffic.
- Treat API responses as OpenAI-compatible chat completion payloads. Do not depend on GonkaGate-specific billing totals in `response.usage`; track completed spend in the dashboard.

## Common errors and limits

- `401 invalid_api_key` usually means the key is missing, malformed, or not loaded in the runtime that creates the client.
- `404 model_not_found` usually means the SDK setup is fine but the model ID is stale.
- `429 insufficient_quota` means the prepaid USD balance is too low for the request.

## See also

- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, secure storage, and rotation.
- [OpenAI SDK compatibility](/en/docs/sdk/openai) for the shared GonkaGate rules that stay the same across runtimes.
- [API Reference Overview](/en/docs/api/reference/overview) when this TypeScript request works and you need exact request fields, streaming behavior, or runtime-independent failure policy.
- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you are switching already-working OpenAI-compatible code.
- [Framework & Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview) if you use Vercel AI SDK, LangChain, or another wrapper instead of the raw OpenAI client.
- [Claude Code](/en/docs/guides/coding-agents/claude-code), [Cursor](/en/docs/guides/coding-agents/cursor), [Kilo Code](/en/docs/guides/coding-agents/kilo-code), [MiMoCode](/en/docs/guides/coding-agents/mimo-code), [OpenCode](/en/docs/guides/coding-agents/opencode), [OpenClaw](/en/docs/guides/coding-agents/openclaw), and [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) if your TypeScript workflow runs through an agent-owned tool path.
- [Get Models endpoint reference](/en/docs/api/api-reference/models/get-models) for the current machine-readable model IDs.