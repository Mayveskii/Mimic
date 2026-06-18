# OpenClaw Custom Provider Setup

Set up GonkaGate as an OpenClaw provider.

Use this guide if OpenClaw already owns the provider layer in your stack. The goal here is simple: get one routed request working first, then verify that OpenClaw is actually sending traffic through GonkaGate.

If you only need the direct SDK path, use [OpenAI SDK Compatibility](/en/docs/sdk/openai). If you need guides for wrappers and community tools instead, use [Framework and Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview).

## Fast path

Use this path when the standard OpenClaw onboarding already fits your setup. You need an OpenClaw install, a GonkaGate API key, and a model ID from [Get models](/en/docs/api/api-reference/models/get-models).

### Interactive onboarding
 Installation
```
openclaw onboard
```

### Non-interactive onboarding
 Non-interactive onboarding
```
openclaw onboard --non-interactive \
  --auth-choice custom-api-key \
  --custom-base-url "https://api.gonkagate.com/v1" \
  --custom-model-id "qwen/qwen3-235b-a22b-instruct-2507-fp8" \
  --custom-api-key "$GONKAGATE_API_KEY" \
  --custom-compatibility openai
```

If your installed OpenClaw version exposes different flags, check the [OpenClaw onboard reference](https://docs.openclaw.ai/cli/onboard).

### Verify after onboarding
 Verify after onboarding
```
openclaw models list
openclaw models status
```
 Example
```
/models
/model qwen3-235b
/status
```

Confirm that:

- `qwen3-235b` appears in `/models`.
- `/model qwen3-235b` switches the active model.
- A test prompt returns a valid answer through GonkaGate.

## Manual config

If you need to review or edit the JSON directly, make sure it includes these four things:

- `GONKAGATE_API_KEY` in an env block or another secret source OpenClaw can read.
- A provider entry under `models.providers.gonkagate`.
- The exact GonkaGate model ID inside the provider `models[]` list.
- A fully-qualified allowlist key under `agents.defaults.models`.

Provider definitions usually live in `~/.openclaw/openclaw.json`. Older installs may still point to `~/.clawdbot/clawdbot.json`, usually as a symlink.
 Configuration
```
{
  "env": {
    "GONKAGATE_API_KEY": "gp-..."
  },
  "models": {
    "mode": "merge",
    "providers": {
      "gonkagate": {
        "baseUrl": "https://api.gonkagate.com/v1",
        "apiKey": "${GONKAGATE_API_KEY}",
        "api": "openai-completions",
        "models": [
          {
            "id": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
            "name": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
            "reasoning": false,
            "input": ["text"],
            "cost": {
              "input": 0,
              "output": 0,
              "cacheRead": 0,
              "cacheWrite": 0
            },
            "contextWindow": 262144,
            "maxTokens": 8192
          }
        ]
      }
    }
  },
  "agents": {
    "defaults": {
      "models": {
        "gonkagate/qwen/qwen3-235b-a22b-instruct-2507-fp8": {
          "alias": "qwen3-235b"
        }
      },
      "model": {
        "primary": "gonkagate/qwen/qwen3-235b-a22b-instruct-2507-fp8"
      }
    }
  }
}
```

Keep these fields straight:

- `api` must be `openai-completions` so OpenClaw sends requests to the OpenAI-compatible `/chat/completions` path.
- `baseUrl` must be `https://api.gonkagate.com/v1`.
- The model `id` inside the provider must match the exact GonkaGate model ID.
- The shorter alias belongs only in `agents.defaults.models`, not in the provider `id`.

## Apply and verify

Apply the edited config so OpenClaw reloads the provider metadata and allowlist.
 Configuration
```
openclaw gateway config.apply --file ~/.openclaw/openclaw.json
```

If your build does not support `config.apply` yet, restart the gateway instead:
 Configuration
```
openclaw gateway restart
```

Then run both CLI and chat-level checks:

- `openclaw models list` to confirm the provider metadata loaded.
- `openclaw models status` to confirm the gateway sees the provider.
- `/models` to confirm the alias is visible.
- `/model qwen3-235b` plus `/status` to confirm the selected model changed.
- One test prompt to confirm a full request path through GonkaGate.

## Common failures

### `model not allowed`

The allowlist entry is missing or malformed.

1. Use the fully-qualified key `gonkagate/qwen/qwen3-235b-a22b-instruct-2507-fp8`.
2. Keep `qwen3-235b` only as the alias value.
3. Re-apply the config after editing JSON.

### The model does not appear in `/models`

Usually one half is missing: provider metadata or allowlist entry.

1. Confirm `models.providers.gonkagate.models[]` contains the model ID.
2. Confirm `agents.defaults.models` contains the same fully-qualified key.
3. Apply the config again and re-run `/models`.

### The wrong model gets called

OpenClaw can only forward the model ID you put in the provider config.

1. Set `id` exactly to `qwen/qwen3-235b-a22b-instruct-2507-fp8`.
2. Do not shorten or rename that ID in the provider block.
3. Keep the alias only in the allowlist entry.

### Auth or base URL errors

Check the shared connection values before you spend more time debugging OpenClaw itself.

1. Confirm the base URL is `https://api.gonkagate.com/v1`.
2. Confirm the API key starts with `gp-`.
3. Check for JSON syntax mistakes before re-applying the config.

## Related pages

- [Claude Code](/en/docs/guides/coding-agents/claude-code) for the supported Anthropic-compatible setup
- [Cursor setup](/en/docs/guides/coding-agents/cursor) for the OpenAI settings path
- [Kilo Code](/en/docs/guides/coding-agents/kilo-code) for the Kilo Code installer path
- [MiMoCode](/en/docs/guides/coding-agents/mimo-code) for the MiMoCode installer path
- [OpenCode setup](/en/docs/guides/coding-agents/opencode) for another custom-provider path
- [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) for the Hermes installer path
- [Get models](/en/docs/api/api-reference/models/get-models) for current machine-readable model IDs before editing provider config
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, header format, and rotation
- [API Reference Overview](/en/docs/api/reference/overview) for the exact request and response contract after OpenClaw forwards the call