# Cursor Setup

Connect Cursor to GonkaGate through Cursor’s OpenAI settings.

Connect Cursor to GonkaGate by turning on its OpenAI settings, setting the base URL, and adding one live model ID. Then send one short chat request before you trust it in a longer session.

## Set up Cursor

Open `Settings -> Cursor Settings -> Models`, then configure:

- In `API Keys`, enable the `OpenAI API Key` toggle, then paste `gp-your-api-key`
- `Override OpenAI Base URL`: enabled
- `Base URL`: `https://api.gonkagate.com/v1`
- `Add or search model`: add an exact GonkaGate model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models), for example `qwen/qwen3-235b-a22b-instruct-2507-fp8`

After you add the model, make sure it is enabled in Cursor’s models list before you try to use it in chat.

## Before you start

- Cursor is already installed.
- You already have a GonkaGate API key in `gp-...` format.
- This page covers the direct Cursor setup, not a broader SDK or migration path.
- You will choose a current live model ID from GonkaGate instead of assuming OpenAI default model names.

## Verify the connection in Cursor

Open the Cursor chat panel, select the model you added, and send this prompt:

`Reply with exactly: Cursor connected to GonkaGate`

That confirms the API key, base URL override, and model selection before a longer coding session.

## What usually trips people up in Cursor

- Keep `/v1` in the base URL. Cursor can save the setting even when the endpoint is wrong.
- The `OpenAI API Key` switch must stay enabled. A filled input field alone is not enough.
- Add the exact live GonkaGate model ID manually if it does not appear automatically in the search box.
- Adding a model is not enough. It also has to stay enabled in the settings list, or it will not appear in chat.
- One short chat prompt after setup is safer than trusting the settings screen alone.

## Common first failures

| If you see | What it usually means | What to do |
|---|---|---|
| Requests fail immediately or verification does not work | The base URL is wrong or the API key is not the GonkaGate key you intended to use | Recheck https://api.gonkagate.com/v1, then confirm the API key |
| The key looks saved but Cursor still behaves as if no provider is configured | The OpenAI API Key toggle is off in API Keys | Turn the toggle on, not just the input field |
| The model does not appear in chat | The model ID is wrong, not enabled, or the picker is stale | Verify the model ID in GET /v1/models, add it again if needed, ensure it is enabled, then restart Cursor if necessary |
| 401 invalid_api_key | The key is missing, invalid, or pasted into the wrong provider field | Recheck the key and Authentication and API Keys |
| 404 model_not_found | The model ID is stale or invalid | Switch to an exact current model ID from GET /v1/models |
| 429 insufficient_quota | The prepaid USD balance is too low for this request | Check Pricing before retrying |
| 429 rate_limit_exceeded | The request was throttled | Retry with bounded backoff and use Rate Limit Handling for the full policy |

## See also

- [Claude Code](/en/docs/guides/coding-agents/claude-code) for the supported Anthropic-compatible path
- [Kilo Code](/en/docs/guides/coding-agents/kilo-code) for the Kilo Code installer path
- [MiMoCode](/en/docs/guides/coding-agents/mimo-code) for the MiMoCode installer path
- [OpenCode setup](/en/docs/guides/coding-agents/opencode) for the OpenCode installer path
- [OpenClaw setup](/en/docs/guides/coding-agents/openclaw) for the custom-provider route
- [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) for the Hermes installer path
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key storage, rotation, and auth troubleshooting
- [Model Selection Guide](/en/docs/guides/overview/models) for choosing a model before you keep it in Cursor
- [GET /v1/models](/en/docs/api/api-reference/models/get-models) to confirm the exact live model ID that Cursor should use