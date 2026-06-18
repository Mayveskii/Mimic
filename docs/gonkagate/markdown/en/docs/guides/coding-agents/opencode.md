# OpenCode

Connect OpenCode to GonkaGate with the official installer or the manual custom-provider path.

Connect an existing OpenCode install to GonkaGate. Start with the official installer unless you specifically want to inspect or manage the provider config yourself.

## Quick setup with the GonkaGate installer

Run the official installer:
 Quick setup with the GonkaGate installer
```
npx @gonkagate/opencode-setup
```

This tool configures an existing local `opencode` install. It does not install OpenCode itself.

Use this path if you want the safest default setup without hand-editing OpenCode config:

- keeps the secret out of repo-local `opencode.json`
- writes only the minimum safe config
- checks the effective result with `opencode debug config --pure`
- leaves you with plain `opencode`

## Before you start

- OpenCode is already installed locally and available on `PATH`.
- Node `>=22.14.0` is available for the `npx` installer.
- You already have a GonkaGate API key in `gp-...` format.
- The installer currently starts with `qwen/qwen3-235b-a22b-instruct-2507-fp8`.

## Non-interactive installer path

Use this for scripts, automation, or if you want to bind a repo in project scope:
 Non-interactive installer path
```
GONKAGATE_API_KEY=gp-... npx @gonkagate/opencode-setup --scope project --yes
```

If you do not want to pass the key through an environment variable, use stdin:
 Environment Variables
```
printf '%s' "$GONKAGATE_API_KEY" | npx @gonkagate/opencode-setup --api-key-stdin --scope project --yes --json
```

## Verify the installer path

Run:
 Verify the installer path
```
opencode debug config --pure
```

Check that the resolved config includes:

- `provider.gonkagate`
- `https://api.gonkagate.com/v1`
- `qwen/qwen3-235b-a22b-instruct-2507-fp8` as the starting GonkaGate model

Then open OpenCode and:

- run `/models`
- pick the GonkaGate model if needed
- send this prompt:

`Reply with exactly: OpenCode connected to GonkaGate`

That confirms the provider, base URL, and model routing before a longer session.

## Manual custom-provider path

Use this path only if you want to manage the OpenCode provider config yourself or compare the installer result with OpenCode’s documented custom-provider flow.

### 1. Store the credential in OpenCode

Start OpenCode, run `/connect`, choose `Other`, then enter:

- Provider id: `gonkagate`
- API key: your `gp-...` key

OpenCode stores `/connect` credentials in `~/.local/share/opencode/auth.json`.

### 2. Add the provider config

OpenCode reads global config from `~/.config/opencode/opencode.json` and can also read a project `./opencode.json`. Project config has higher precedence.
 Configuration
```
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "gonkagate": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "GonkaGate",
      "options": {
        "baseURL": "https://api.gonkagate.com/v1"
      },
      "models": {
        "qwen/qwen3-235b-a22b-instruct-2507-fp8": {
          "name": "GonkaGate Qwen3 235B"
        }
      }
    }
  },
  "model": "gonkagate/qwen/qwen3-235b-a22b-instruct-2507-fp8",
  "small_model": "gonkagate/qwen/qwen3-235b-a22b-instruct-2507-fp8"
}
```

Keep these points straight:

- Use `@ai-sdk/openai-compatible` because the current GonkaGate OpenCode path is the OpenAI-compatible `chat/completions` flow.
- Keep `/v1` in the `baseURL`.
- The provider id in `/connect` and the provider key in `opencode.json` must match exactly.
- Start with `qwen/qwen3-235b-a22b-instruct-2507-fp8`, then switch only after you confirm another current model ID in [Get Models](/en/docs/api/api-reference/models/get-models).

If you prefer environment-based auth instead of `/connect`, OpenCode also supports `options.apiKey`, for example `"{env:GONKAGATE_API_KEY}"`. Keep that value out of repo-local config.

### 3. Verify the manual path

Run:
 3. Verify the manual path
```
opencode auth list
opencode debug config --pure
```

Then start OpenCode and run:
 Then start OpenCode and run
```
/models
```

Confirm that `gonkagate` and your configured model appear before you send a real prompt.

## Common first failures

| If you see | What it usually means | What to do |
|---|---|---|
| gonkagate does not appear in /models | The provider id does not match, or the provider is blocked | Make sure /connect and opencode.json both use gonkagate, then check enabled_providers and disabled_providers |
| opencode debug config --pure does not resolve GonkaGate | A higher-precedence config layer overrides this setup | Recheck project opencode.json, OPENCODE_CONFIG, and any existing user config |
| 401 invalid_api_key | The gp-... key is missing, invalid, or not bound where OpenCode expects it | Recheck the installer/manual auth step and your key |
| 404 model_not_found | The model ID is stale | Refresh it from Get Models and update the config |
| The installer completed but the current session still uses another provider | The durable config is correct but the current session is overridden or blocked | Re-run npx @gonkagate/opencode-setup, then compare the current result with opencode debug config --pure |

## Current limits

- This guide configures an existing OpenCode install. It does not install OpenCode itself.
- The current verified GonkaGate OpenCode path is the OpenAI-compatible `/chat/completions` flow.
- The installer starts with `qwen/qwen3-235b-a22b-instruct-2507-fp8`.
- Project `opencode.json` has higher precedence than global config, so local overrides can hide a correct user-level provider setup.

## Official references

- [GonkaGate OpenCode installer](https://github.com/GonkaGate/opencode-setup)
- [OpenCode providers docs](https://opencode.ai/docs/providers/)
- [OpenCode config docs](https://opencode.ai/docs/config/)

## See also

- [Claude Code](/en/docs/guides/coding-agents/claude-code) for the supported Anthropic-compatible path
- [Cursor setup](/en/docs/guides/coding-agents/cursor) for Cursor’s OpenAI settings path
- [Kilo Code](/en/docs/guides/coding-agents/kilo-code) for the Kilo Code installer path
- [MiMoCode](/en/docs/guides/coding-agents/mimo-code) for the MiMoCode installer path
- [OpenClaw setup](/en/docs/guides/coding-agents/openclaw) for another custom-provider route
- [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) for the Hermes installer path
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, storage, and rotation
- [Get Models](/en/docs/api/api-reference/models/get-models) for current machine-readable model IDs
- [GonkaGate Quickstart](/en/docs/quickstart) if you still want one direct request from scratch