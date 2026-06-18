# Aider Setup

Connect Aider to GonkaGate in a local repo.

Connect Aider to GonkaGate by exporting `OPENAI_API_KEY`, setting the GonkaGate base URL, and using `openai/<model-id>`. Start one session first or save the same values in `.aider.conf.yml`, then verify the setup before real edits.

## Start one Aider session

Export your GonkaGate key through `OPENAI_API_KEY`, then start Aider with the GonkaGate base URL and an `openai/<model-id>` value.
**Endpoint**
```
export OPENAI_API_KEY="gp-your-api-key"
cd /path/to/your/repo

aider \
  --openai-api-base https://api.gonkagate.com/v1 \
  --model openai/qwen/qwen3-235b-a22b-instruct-2507-fp8
```

Replace the example model with a current ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models).

## Before you start

- Aider is already installed.
- You are working in a local project directory or git repo.
- You already have a GonkaGate API key in `gp-...` format.
- You need the narrow Aider setup path, not a broader migration or SDK walkthrough.

## Save the setup in the repo

Use a repo-local config when the same project should always open Aider against GonkaGate.

Create `.aider.conf.yml` in the repo root:
 Save the setup in the repo
```
model: openai/qwen/qwen3-235b-a22b-instruct-2507-fp8
openai-api-base: https://api.gonkagate.com/v1
```

Then start Aider from that repo:
 Then start Aider from that repo
```
export OPENAI_API_KEY="gp-your-api-key"
cd /path/to/your/repo
aider
```

Aider can load config from the current directory, the repo root, or your home directory. Prefer a repo-local config when GonkaGate should apply only to one project.

## Aider-specific rules

- Keep the `openai/` prefix in Aider model names.
- The part after `openai/` must match a current GonkaGate model ID.
- Aider reads the API key from `OPENAI_API_KEY`, so the key must be available in the shell where you launch Aider.

## Verify inside your repo

Launch Aider from the repo you want to work on and send this prompt in chat:

`Reply with exactly: Aider connected to GonkaGate`

That confirms the endpoint, API key, and model path before you start real file edits.

## Common first failures

| If you see | What it usually means | What to do |
|---|---|---|
| model not found | The openai/ prefix is missing or the model ID is stale | Keep the openai/ prefix and verify the model ID in GET /v1/models |
| Unknown context-window or token-cost warnings | Aider does not recognize the model metadata yet | The endpoint can still work, but validate the model before real edits |
| 401 invalid_api_key | The key is missing, invalid, or loaded in the wrong shell | Recheck OPENAI_API_KEY and Authentication and API Keys |
| 429 insufficient_quota | The prepaid USD balance is too low for this request | Check Pricing before retrying |
| 429 rate_limit_exceeded | The request was throttled | Retry with bounded backoff and use Rate Limit Handling for the full policy |

## Official references

- [Aider OpenAI-compatible docs](https://aider.chat/docs/llms/openai-compat.html)
- [Aider YAML config docs](https://aider.chat/docs/config/aider_conf.html)
- [Aider model warnings docs](https://aider.chat/docs/troubleshooting/warnings.html)

Last verified: 2026-03-13 against the official Aider OpenAI-compatible and configuration docs.

## See also

- [Framework and Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview) for a different wrapper or coding-tool path
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key storage, rotation, and auth troubleshooting
- [Model Selection Guide](/en/docs/guides/overview/models) for choosing a model before you lock it into Aider
- [GET /v1/models](/en/docs/api/api-reference/models/get-models) to confirm the exact live model ID that Aider should use