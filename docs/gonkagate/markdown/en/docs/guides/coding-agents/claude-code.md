# Claude Code

Connect Claude Code to GonkaGate through the supported Anthropic-compatible gateway.

Connect an existing Claude Code install to GonkaGate through the supported Anthropic-compatible gateway. Start with the official installer. Then check the connection before you rely on it in a longer session.

## Quick setup

Run the official installer:
 Quick setup
```
npx @gonkagate/claude-code
```

This tool configures an existing local `Claude Code` install. It does not install `Claude Code` itself.

Use a regular GonkaGate API key in `gp-...` format. Point Claude Code at the deployment root:
 Example
```
ANTHROPIC_BASE_URL=https://api.gonkagate.com
```

Do not append `/v1`. Claude Code resolves Anthropic-compatible requests through `POST /v1/messages` and `POST /v1/messages/count_tokens`.

If Claude Code was previously connected directly to Anthropic, run this once before you relaunch it:
 Command
```
claude auth logout
```

## Before you start

- Claude Code is already installed locally.
- You already have a GonkaGate API key in `gp-...` format.
- The installer currently uses `qwen/qwen3-235b-a22b-instruct-2507-fp8`.

## Verify the setup

Start Claude Code:
 Verify the setup
```
claude
```

Then:

- run `/status`
- confirm that the active gateway is `https://api.gonkagate.com`
- confirm that the active model is `qwen/qwen3-235b-a22b-instruct-2507-fp8`

If you want to check the gateway without opening Claude Code first, send a token-count request:
**Request Example**
```
curl https://api.gonkagate.com/v1/messages/count_tokens \
  -H "Authorization: Bearer gp-YOUR_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    "messages": [
      {
        "role": "user",
        "content": [{"type": "text", "text": "Hello"}]
      }
    ]
  }'
```

Expected result: `200 OK` with a JSON body that includes `input_tokens`.

## Current limits

- This guide covers only the supported Claude Code path on GonkaGate.
- The current public Claude Code model is `qwen/qwen3-235b-a22b-instruct-2507-fp8`.
- Use `ANTHROPIC_BASE_URL=https://api.gonkagate.com` at the deployment root, not `https://api.gonkagate.com/v1`.
- Do not assume arbitrary Anthropic model IDs, full Anthropic API coverage, or custom base URL/model flexibility for this setup.

## Official references

- [GonkaGate Claude Code installer](https://github.com/GonkaGate/gonkagate-claude-code)

## See also

- [Cursor setup](/en/docs/guides/coding-agents/cursor) for Cursor’s OpenAI settings path
- [Kilo Code](/en/docs/guides/coding-agents/kilo-code) for the Kilo Code installer path
- [MiMoCode](/en/docs/guides/coding-agents/mimo-code) for the MiMoCode installer path
- [OpenCode setup](/en/docs/guides/coding-agents/opencode) for the OpenCode installer path
- [OpenClaw setup](/en/docs/guides/coding-agents/openclaw) for the custom-provider route
- [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) for the Hermes installer path
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, storage, and rotation
- [GonkaGate Quickstart](/en/docs/quickstart) if you still need one verified request from scratch