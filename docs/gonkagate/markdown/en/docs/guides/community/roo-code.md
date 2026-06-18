# Roo Code Setup

Connect Roo Code to GonkaGate.

Connect Roo Code to GonkaGate through Roo’s OpenAI-compatible provider, then prove the setup in two steps: one plain-text reply and one read-only workspace task.

## Set the provider values

Open Roo Code settings and set:

- API Provider: `OpenAI Compatible`
- Base URL: `https://api.gonkagate.com/v1`
- API Key: `gp-your-api-key`
- Model ID: use a model from [GET /v1/models](/en/docs/api/api-reference/models/get-models)

If Roo asks for advanced model configuration, treat that as Roo-specific tuning. Start with the working model ID first, then adjust optional settings only after the connection succeeds.

## Before you start

- Roo Code is already installed.
- You already have a GonkaGate API key in `gp-...` format.
- You need the narrow Roo provider path, not a broader migration or raw SDK walkthrough.
- For real Roo workflows, you will separately validate a model and provider pair that can handle tool-driven tasks.

## Verify the basic connection

Send a short prompt in Roo Code:

`Reply with exactly: Roo Code connected to GonkaGate`

That confirms the base URL, API key, and model ID are wired correctly.

## Validate one read-only workspace task

Roo Code uses native tool calling in real coding sessions. Before you trust the setup for actual edits, run one read-only workspace prompt such as:

`List the files in the current workspace and summarize the repo root without changing anything.`

Treat that as a first canary before a real coding session, not as a blanket guarantee for every model or every tool-heavy workflow.

## Roo-specific rules

- `GET /v1/models` gives you valid GonkaGate model IDs, but it is not a tool-support contract by itself.
- A plain text reply proves the connection path works. It does not prove that a tool-driven Roo workflow is ready.
- If you need Roo for real workspace tasks, validate the model against the [live models catalog](/en/models) and the [Tool Calling guide](/en/docs/guides/features/tool-calling) before you blame the provider setup.

## Common first failures

| If you see | What it usually means | What to do |
|---|---|---|
| 401 invalid_api_key | The key is missing, invalid, or pasted into the wrong provider profile | Recheck the key and Authentication and API Keys |
| 404 or model not found | The model ID does not match a current GonkaGate model from GET /v1/models | Refresh the model ID before treating the provider setup as broken |
| 429 insufficient_quota | The prepaid USD balance is too low for this request | Check Pricing before retrying |
| 429 rate_limit_exceeded | The request was throttled | Retry with bounded backoff and use Rate Limit Handling for the full policy |
| Text prompts work but workspace tasks fail | The connection path works, but the selected model/provider pair is not ready for tool-driven Roo tasks | Recheck the model against the live models catalog and the Tool Calling guide before blaming the provider setup |

## Official references

- [Roo Code OpenAI-compatible provider docs](https://docs.roocode.com/providers/openai-compatible)

Last verified: 2026-03-13 against the official Roo Code OpenAI-compatible provider docs.

## See also

- [Framework and Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview) for a different framework or coding-tool path
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key storage, rotation, and auth troubleshooting
- [GET /v1/models](/en/docs/api/api-reference/models/get-models) to confirm the exact live model ID that Roo should use
- [Tool Calling guide](/en/docs/guides/features/tool-calling) to validate whether the selected model/provider pair is good enough for Roo workspace tasks