# MiMoCode

Connect MiMoCode to GonkaGate with the official mimo-code-setup installer.

Connect an existing MiMoCode install to GonkaGate with the official installer. This guide uses the supported `@gonkagate/mimo-code-setup` path, so you do not have to edit MiMoCode config by hand or put secrets in shell profiles or repository files.

## Quick setup with the official installer

Run the official installer:
 Quick setup with the official installer
```
npx @gonkagate/mimo-code-setup
```

This tool configures an existing local `mimo` install. It does not install MiMoCode itself.

Use it if you want the safe default setup without hand-editing MiMoCode config:

- keeps the secret out of repository-local config
- preserves unrelated MiMoCode settings instead of rewriting whole files
- picks the recommended scope for the current directory
- verifies the resolved plain-`mimo` result after writing config
- leaves you with plain `mimo`

## Before you start

- MiMoCode is already installed locally and available as `mimo` on `PATH`.
- MiMoCode is `0.1.0` or newer.
- Node `>=22.14.0` is available for the `npx` installer.
- You already have a GonkaGate API key in `gp-...` format.
- The installer currently starts with `moonshotai/kimi-k2.6`.
- The supported platforms are macOS, Linux, native Windows, and WSL.

## Choose the right scope

The installer supports two scopes:

- `user`: use GonkaGate on this machine in general.
- `project`: make GonkaGate the default for the current repository.

By default, the interactive flow usually picks:

- inside a git repository: `project`
- outside a repository: `user`

One limit matters here: `project` scope is commit-safe because `.mimocode/mimocode.json` stores only activation settings. The actual `provider.gonkagate` definition and secret binding still live in user config, so every machine that uses the repository still needs to run the installer once.

## Non-interactive setup

Use this for scripts, automation, or repeatable local setup:
 Non-interactive setup
```
npx @gonkagate/mimo-code-setup --scope project --yes
```

With a key from the environment:
 With a key from the environment
```
GONKAGATE_API_KEY="$GONKAGATE_API_KEY" npx @gonkagate/mimo-code-setup --scope user --yes
```

With a key through stdin and JSON output:
 With a key through stdin and JSON output
```
printf '%s' "$GONKAGATE_API_KEY" | npx @gonkagate/mimo-code-setup --api-key-stdin --scope project --yes --json
```

Pass the key only through:

- the hidden interactive prompt
- `GONKAGATE_API_KEY`
- `--api-key-stdin`

The installer intentionally does not accept plain `--api-key`, so secrets do not end up in shell history or process lists.

## Verify the setup

If MiMoCode is already open, close it first. Then start a fresh session:
 Verify the setup
```
mimo
```

Send one short prompt before a longer coding session:

`Reply with exactly: MiMoCode connected to GonkaGate`

The installer verifies the resolved MiMoCode config and model visibility before it finishes. That does not prove billing, quota, or full first-request readiness for the first billable request.

If you need config proof outside the chat session, use redacted installer diagnostics. Do not paste raw `mimo --pure debug config` output into support requests because resolved config can include substituted secret values.

## Where mimo-code-setup writes data

The installer keeps secrets out of the repository by default. It writes or manages these locations:

- user-level durable config: `~/.config/mimocode/mimocode.json`
- created global config filename when no existing MiMoCode config target is present: `mimocode.jsonc`
- project-level durable config: `.mimocode/mimocode.json`
- user-level secret: `~/.gonkagate/mimo-code/api-key`
- install state: `~/.gonkagate/mimo-code/install-state.json`
- project-config rollback backups: `~/.gonkagate/mimo-code/backups/project-config`

How the writes are split:

- `user` scope writes provider and activation into user config
- `project` scope keeps the provider definition and secret binding in user config, then writes only activation into project config

The canonical secret binding is `provider.gonkagate.options.apiKey = {file:~/.gonkagate/mimo-code/api-key}`. It belongs in user config, not in repository-local `.mimocode/mimocode.json`.

## Common first failures

| If you see | What it usually means | What to do |
|---|---|---|
| Setup stops before any files are written | mimo is missing, or the installed MiMoCode version is older than the supported 0.1.0 baseline | Fix PATH or update MiMoCode, then rerun the installer |
| Node.js is below 22.14.0 | The npx runtime is too old for the installer | Upgrade Node.js and rerun the helper |
| The installer says the saved setup succeeded, but the current shell still acts wrong | A higher-precedence runtime layer such as MIMOCODE_CONFIG, MIMOCODE_CONFIG_CONTENT, MIMOCODE_CONFIG_DIR, MIMOCODE_AUTH_CONTENT, or MIMOCODE_DISABLE_PROJECT_CONFIG overrides it | Exit the active MiMoCode session, start plain mimo, and recheck the override variables |
| project scope works on one machine but not another | Project scope activates GonkaGate per repo, but the user-level provider definition and secret remain machine-local | Run the installer on each participating machine |
| You want to pass --api-key directly | Plain --api-key is intentionally unsupported | Use the hidden prompt, GONKAGATE_API_KEY, or --api-key-stdin instead |
| First MiMoCode prompt fails after setup | Config verification proved provider wiring and model visibility, not balance or first-request readiness | Check balance, quota, and the selected model before retrying |

## Current limits

- This guide configures an existing MiMoCode install. It does not install MiMoCode itself.
- The current public setup model is `moonshotai/kimi-k2.6`.
- The current transport target is OpenAI-compatible `chat_completions` through `https://api.gonkagate.com/v1`.
- The stable provider id is `gonkagate`, using the managed `provider.gonkagate` config.
- The installer does not accept arbitrary custom base URLs in v1.
- The installer does not accept arbitrary model IDs in v1.
- This page does not claim `/v1/responses` support.
- The installer does not mutate shell profiles, generate `.env` files, or store secrets in repository-local config.

## Official references

- [GonkaGate MiMoCode installer](https://github.com/GonkaGate/mimo-code-setup)
- [MiMoCode installer implementation notes](https://github.com/GonkaGate/mimo-code-setup/blob/main/docs/how-it-works.md)
- [MiMoCode installer security notes](https://github.com/GonkaGate/mimo-code-setup/blob/main/docs/security.md)

## See also

- [Claude Code](/en/docs/guides/coding-agents/claude-code) for the supported Anthropic-compatible path
- [Cursor setup](/en/docs/guides/coding-agents/cursor) for Cursor’s OpenAI settings path
- [Kilo Code](/en/docs/guides/coding-agents/kilo-code) for the Kilo Code installer path
- [OpenCode setup](/en/docs/guides/coding-agents/opencode) for the OpenCode installer path
- [OpenClaw setup](/en/docs/guides/coding-agents/openclaw) for the custom-provider route
- [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) for the Hermes installer path
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, storage, and rotation
- [Get Models](/en/docs/api/api-reference/models/get-models) for current machine-readable model IDs