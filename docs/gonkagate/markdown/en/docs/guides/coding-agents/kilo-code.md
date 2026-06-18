# Kilo Code

Connect Kilo Code to GonkaGate with the official kilo-setup installer.

Connect an existing Kilo Code install to GonkaGate with the official installer. This guide uses the supported `@gonkagate/kilo-setup` path, so you do not have to edit Kilo config by hand or stash secrets in a repository.

## Quick setup with the official installer

Run the official installer:
 Quick setup with the official installer
```
npx @gonkagate/kilo-setup
```

This tool configures an existing local `kilo` install. It does not install Kilo Code itself.

Use it if you want the safe default setup without hand-editing Kilo config:

- keeps the secret out of config files in the repository by default
- preserves unrelated Kilo settings instead of rewriting whole files
- picks the recommended scope for the current directory
- checks the result after writing config
- leaves you with plain `kilo`

## Before you start

- Kilo Code is already installed locally and available as `kilo`, or the fallback `kilocode` command exists on `PATH`.
- The supported Kilo CLI version is exact `@kilocode/cli@7.2.0`.
- Node `>=22.14.0` is available for the `npx` installer.
- You already have a GonkaGate API key in `gp-...` format.
- The installer currently starts with `qwen/qwen3-235b-a22b-instruct-2507-fp8`.

## Choose the right scope

The installer supports two scopes:

- `user`: use GonkaGate on this machine in general.
- `project`: make GonkaGate the default for the current repository.

By default, the interactive flow picks:

- inside a git repository: `project`
- outside a repository: `user`

One limit matters here: `project` scope is commit-safe because `.kilo/kilo.jsonc` stores only activation settings. The actual `provider.gonkagate` definition still lives in user config, so every machine that uses the repository still needs to run the installer once.

## Non-interactive setup

Use this for scripts, automation, or repeatable local setup:
 Non-interactive setup
```
npx @gonkagate/kilo-setup --scope project --yes
```

With a key from the environment:
 With a key from the environment
```
GONKAGATE_API_KEY="$GONKAGATE_API_KEY" npx @gonkagate/kilo-setup --scope user --yes
```

With a key through stdin and JSON output:
 With a key through stdin and JSON output
```
printf '%s' "$GONKAGATE_API_KEY" | npx @gonkagate/kilo-setup --api-key-stdin --scope project --yes --json
```

If Kilo keeps reopening the last UI-selected model, clear that cache too:
 Command
```
npx @gonkagate/kilo-setup --scope project --clear-kilo-model-cache --yes
```

Pass the key only through:

- the hidden interactive prompt
- `GONKAGATE_API_KEY`
- `--api-key-stdin`

The installer intentionally does not accept plain `--api-key`, so secrets do not end up in shell history or process lists.

## Verify the setup

If Kilo is already open, close it first. Then start a fresh session:
 Verify the setup
```
kilo
```

Then:

- run `/models`
- select the GonkaGate model if it is not active yet
- send this prompt:

`Reply with exactly: Kilo Code connected to GonkaGate`

That is enough to confirm the provider wiring before a longer coding session.

If it still looks unchanged, check `KILO_CONFIG`, `KILO_CONFIG_DIR`, and `KILO_CONFIG_CONTENT` first. Any of them can override the durable config. Also make sure you are not still inside an older Kilo session that kept runtime overrides.

## Where kilo-setup writes data

The installer keeps secrets out of the repository by default. It writes to three places:

- user-level secret: `~/.gonkagate/kilo/api-key`
- user-level durable config: `~/.config/kilo/kilo.jsonc`
- project-level durable config: `.kilo/kilo.jsonc`

How the writes are split:

- `user` scope writes provider and activation into user config
- `project` scope keeps the provider definition and secret binding in user config, then writes only activation into project config

## Common first failures

| If you see | What it usually means | What to do |
|---|---|---|
| Setup stops before any files are written | kilo and kilocode are both missing, or the installed Kilo version is not exact @kilocode/cli@7.2.0 | Fix PATH, then recheck the local Kilo version |
| The installer says the saved setup succeeded, but the current shell still acts wrong | A higher-precedence runtime layer such as KILO_CONFIG, KILO_CONFIG_DIR, or KILO_CONFIG_CONTENT is still overriding the result | Exit the active Kilo session, start plain kilo, and recheck the override variables |
| Another repository still opens on a GonkaGate model | Kilo reused the last model selected in the UI from ~/.local/state/kilo/model.json | Switch the model in plain kilo or rerun the installer with --clear-kilo-model-cache |
| project scope works on one machine but not another | Project scope activates GonkaGate per repo, but the user-level provider definition is still machine-local | Run the installer on each participating machine |
| You want to pass --api-key directly | Plain --api-key is intentionally unsupported | Use the hidden prompt, GONKAGATE_API_KEY, or --api-key-stdin instead |

## Current limits

- This guide configures an existing Kilo Code install. It does not install Kilo Code itself.
- The current public setup path is intentionally scoped to exact `@kilocode/cli@7.2.0`.
- The current transport target is OpenAI-compatible `chat/completions`.
- The installer starts with `qwen/qwen3-235b-a22b-instruct-2507-fp8`.
- Native Windows production support is not claimed yet.

## Official references

- [GonkaGate Kilo installer](https://github.com/GonkaGate/kilo-setup)
- [Kilo installer user guide](https://github.com/GonkaGate/kilo-setup/blob/main/docs/user-guide.md)
- [Kilo installer troubleshooting](https://github.com/GonkaGate/kilo-setup/blob/main/docs/troubleshooting.md)

## See also

- [Claude Code](/en/docs/guides/coding-agents/claude-code) for the supported Anthropic-compatible path
- [Cursor setup](/en/docs/guides/coding-agents/cursor) for Cursor’s OpenAI settings path
- [MiMoCode](/en/docs/guides/coding-agents/mimo-code) for the MiMoCode installer path
- [OpenCode setup](/en/docs/guides/coding-agents/opencode) for the OpenCode installer path
- [OpenClaw setup](/en/docs/guides/coding-agents/openclaw) for the custom-provider route
- [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) for the Hermes installer path
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, storage, and rotation
- [Get Models](/en/docs/api/api-reference/models/get-models) for current machine-readable model IDs if you later switch away from the installer default