# Cline Setup

Connect Cline to GonkaGate through OpenAI Compatible.

Connect the Cline VS Code extension to GonkaGate through the `OpenAI Compatible` provider. Set the provider values, click `Verify`, then run one short prompt before you trust a longer coding session.

## Set the provider values

Open Cline settings from the gear icon, then set:

- API Provider: `OpenAI Compatible`
- Base URL: `https://api.gonkagate.com/v1`
- API Key: `gp-your-api-key`
- Model: `<model-id-from-get-v1-models>`

If the model does not appear in the picker, enter the exact GonkaGate model ID manually from [GET /v1/models](/en/docs/api/api-reference/models/get-models).

## Before you start

- The Cline extension is already installed in VS Code.
- You already have a GonkaGate API key in `gp-...` format.
- You need the narrow Cline provider setup, not a broader SDK, migration, or raw API walkthrough.

## Verify the provider connection

Click `Verify` in the Cline settings panel first.

After that succeeds, run one short prompt in chat:

`Reply with exactly: Cline connected to GonkaGate`

That confirms the provider, base URL, API key, and model path before you start a real coding session.

## Cline-specific rules

- Use the exact model ID returned by GonkaGate. Do not assume OpenAI default model names.
- Cline lets you choose the model from a picker or enter it manually. Manual entry is the safer path if the model is missing from the picker.
- Leave `Model Configuration` alone unless you intentionally need custom context-window, max-output-token, image-support, or computer-use hints for a known model.
- `Verify` checks the connection path, but one real chat prompt is still the safer canary before a long session.

## Common first failures

| If you see | What it usually means | What to do |
|---|---|---|
| Verify fails or you get a connection error | The Base URL is wrong or the machine cannot reach the endpoint | Recheck https://api.gonkagate.com/v1, then confirm network access to the endpoint |
| 401 invalid_api_key | The API key is missing, invalid, or copied from the wrong environment | Recheck the key and Authentication and API Keys |
| 404 model_not_found | The model ID is stale or invalid | Switch to an exact model ID from GET /v1/models |
| The model connects but behaves oddly | Custom Model Configuration hints changed the expected behavior | Reset the custom hints and retry with a plain text request |
| 429 insufficient_quota | The prepaid USD balance is too low for this request | Check Pricing before retrying |
| 429 rate_limit_exceeded | The request was throttled | Retry with bounded backoff and use Rate Limit Handling for the full policy |

## Official references

- [Cline OpenAI-compatible provider docs](https://docs.cline.bot/provider-config/openai-compatible)
- [Cline CLI getting started](https://docs.cline.bot/cline-cli/getting-started)

**Last verified:** 2026-03-13 against the official Cline OpenAI-compatible provider docs and Cline CLI docs.

## See also

- [Framework and Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview) for a different framework or tool path
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key storage, rotation, and auth troubleshooting
- [Model Selection Guide](/en/docs/guides/overview/models) for choosing a model before you keep it in Cline
- [GET /v1/models](/en/docs/api/api-reference/models/get-models) to confirm the exact live model ID that Cline should use