# Hermes Agent

Connect Hermes Agent to GonkaGate with the official hermes-agent-setup installer.

Connect an existing `hermes-agent` install to GonkaGate with the official installer. The helper configures Hermes to use GonkaGate as its `provider: custom` OpenAI-compatible endpoint at `https://api.gonkagate.com/v1`.

## Quick setup with the official installer

Run the installer in an interactive terminal:
 Command
```
npx @gonkagate/hermes-agent-setup
```

If you use a named Hermes profile, pass it explicitly:
 Command
```
npx @gonkagate/hermes-agent-setup --profile work
```

The tool configures an existing local Hermes Agent install. It does not install Hermes itself or replace `hermes setup`.

Use this path if you want the supported GonkaGate setup without hand-editing Hermes files:

- prompts for the GonkaGate key through a hidden terminal prompt
- resolves the active Hermes config, including `--profile <name>`
- checks `GET /v1/models` before writing files
- offers only live models that also have checked-in launch qualification artifacts
- creates same-run backups and rolls back `config.yaml` if the later `.env` write fails

## Before you start

- `hermes-agent` is installed and available on `PATH` as `hermes`.
- Hermes Agent is `v2026.5.16` / `v0.14.0` or newer.
- Node `>=22.14.0` is available for the `npx` installer.
- You already have a GonkaGate API key in `gp-...` format.
- You can run the helper in an interactive TTY.
- Your platform is Linux, macOS, or WSL2.

Public onboarding is not positioned for users or entities in the United States of America or U.S. territories.

## What the installer changes

The helper manages these Hermes files:

- `~/.hermes/config.yaml`
- `~/.hermes/.env`

It writes only the GonkaGate-managed surface:

- `model.provider = custom`
- `model.base_url = https://api.gonkagate.com/v1`
- `model.default = <selected qualified GonkaGate model>`
- `OPENAI_API_KEY = <your GonkaGate API key>`

The GonkaGate key is stored only in the resolved Hermes `.env` file. It is not written to `config.yaml` and should not be committed to a repository.

## Verify after setup

When the helper finishes, start Hermes in the same resolved context:
 Verify after setup
```
hermes
```

Send one short prompt before a longer coding session, for example:

`Reply with exactly: Hermes Agent connected to GonkaGate`

The installer’s live `/v1/models` check confirms auth and catalog visibility only. It does not prove billing, quota, or full Hermes runtime readiness for the first billable request.

## Common first failures

| If you see | What it usually means | What to do |
|---|---|---|
| Hermes Agent was not found on PATH | The helper cannot run hermes or resolve the active config context | Install Hermes Agent or fix PATH, then rerun the helper |
| Hermes is below v2026.5.16 / v0.14.0 | The local Hermes contract is older than the supported latest-only setup | Upgrade Hermes Agent, then rerun npx @gonkagate/hermes-agent-setup |
| Node.js is below 22.14.0 | The npx runtime is too old for the installer | Upgrade Node.js and rerun the helper |
| A TTY is required | The helper cannot show the hidden API key prompt | Run the command in an interactive terminal |
| 401 invalid_api_key or catalog/auth failure before any write | The gp-... key is missing, invalid, or cannot see the model catalog | Recheck the key, then confirm it works with GET /v1/models |
| Conflict blockers mention providers, custom_providers, or auth.json | Existing Hermes credential or provider state would make the managed write ambiguous | Resolve the conflicting Hermes-owned config manually, then rerun the helper |
| First Hermes prompt fails after setup | /v1/models proved auth/catalog visibility, not balance or first-request readiness | Check balance, quota, and the selected model before retrying |

## Current limits

- This guide configures an existing Hermes Agent install. It does not install Hermes itself.
- The public helper supports the interactive flow and optional `--profile <name>`. It does not expose `--yes`, `--json`, `--api-key`, or `--api-key-stdin`.
- The current target is OpenAI-compatible `chat/completions` through `https://api.gonkagate.com/v1`.
- The helper does not accept arbitrary custom base URLs.
- The helper does not mutate shell profiles, repository-local `.env` files, or `auth.json` credential pools.
- Native Windows production support is not claimed yet.
- The setup check does not prove billing, quota, or full first-request readiness beyond live `/v1/models` auth and catalog visibility.

## Official references

- [GonkaGate Hermes Agent installer](https://github.com/GonkaGate/hermes-agent-setup)
- [Hermes installer security notes](https://github.com/GonkaGate/hermes-agent-setup/blob/main/docs/security.md)
- [Hermes installer implementation notes](https://github.com/GonkaGate/hermes-agent-setup/blob/main/docs/how-it-works.md)

## See also

- [Claude Code](/en/docs/guides/coding-agents/claude-code) for the supported Anthropic-compatible path
- [Cursor setup](/en/docs/guides/coding-agents/cursor) for Cursor’s OpenAI settings path
- [Kilo Code](/en/docs/guides/coding-agents/kilo-code) for the Kilo Code installer path
- [MiMoCode](/en/docs/guides/coding-agents/mimo-code) for the MiMoCode installer path
- [OpenCode setup](/en/docs/guides/coding-agents/opencode) for the OpenCode installer path
- [OpenClaw setup](/en/docs/guides/coding-agents/openclaw) for the custom-provider route
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, storage, and rotation
- [Get Models](/en/docs/api/api-reference/models/get-models) for current machine-readable model IDs