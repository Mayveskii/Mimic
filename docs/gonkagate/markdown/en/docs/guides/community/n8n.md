# n8n Setup

Connect self-hosted n8n to GonkaGate with the official community node package.

Connect self-hosted `n8n` to GonkaGate with the official community node package `@gonkagate/n8n-nodes-gonkagate`. The fastest path is Community Nodes UI, then the root `GonkaGate` node for `List Models` and `Chat Completion`. Move to `GonkaGate Chat Model` only after the root node already works.

If you only need shared OpenAI-compatible API values outside `n8n`, use [OpenAI SDK Compatibility](/en/docs/sdk/openai) instead.

## What this package gives you

- The root node `GonkaGate`
- The additive AI model node `GonkaGate Chat Model`
- The shared credential `GonkaGate API`
- A GonkaGate-first path for `GET /v1/models` and `POST /v1/chat/completions`

The canonical base URL is fixed to `https://api.gonkagate.com/v1`. In the common `n8n` path, you usually only paste the API key into `GonkaGate API`.

## Before you start

- Use self-hosted `n8n`. This package does not currently promise `n8n` Cloud availability.
- Have a GonkaGate API key in `gp-...` format.
- Keep your `n8n` user data directory persistent if you run Docker.
- Pin an exact npm version or Docker tag for production instead of relying on `latest`.

## Choose your install path

| Path | Use it when | What to do |
|---|---|---|
| Community Nodes UI | You already run normal self-hosted n8n with UI access | Install @gonkagate/n8n-nodes-gonkagate from Settings -> Community Nodes |
| Manual npm install | You manage the host or running container directly | Install the package into the n8n nodes folder, then restart n8n |
| Published Docker image | You want GonkaGate preinstalled in Docker | Run ghcr.io/gonkagate/n8n-nodes-gonkagate or copy the public Compose example |

## Fastest install: Community Nodes UI

Use this when the package is already published to npm and your self-hosted `n8n` instance exposes the normal owner/admin UI.

1. Open your self-hosted `n8n` instance.
2. Open `Settings -> Community Nodes`.
3. Click `Install`.
4. Enter one of these package values:

 Installation
```
@gonkagate/n8n-nodes-gonkagate
@gonkagate/n8n-nodes-gonkagate@<version>
```

1. Confirm the community-node prompt if `n8n` shows it.
2. Wait for installation to finish.
3. Restart `n8n` if your deployment model requires it.
4. Open the node picker and search for `GonkaGate`.

## Manual `npm` install

Use this when you manage the host or a running container directly and want a shell-first install path.

### Host-based `n8n`
 Command
```
mkdir -p ~/.n8n/nodes
cd ~/.n8n/nodes
npm install @gonkagate/n8n-nodes-gonkagate@<version>
```

If you use a custom `N8N_USER_FOLDER`, install into `$N8N_USER_FOLDER/nodes` instead. Then restart `n8n` and search for `GonkaGate` in the node picker.

### Running Docker container

Install into the normal container user folder, then restart the container:
 Installation
```
docker exec -it n8n sh
mkdir -p /home/node/.n8n/nodes
cd /home/node/.n8n/nodes
npm install @gonkagate/n8n-nodes-gonkagate@<version>
exit
docker restart n8n
```

## Docker path

Use the published image when you want the lightest Docker setup with GonkaGate already installed.

Published image:
 Docker path
```
ghcr.io/gonkagate/n8n-nodes-gonkagate
```

Example `docker run` path:
 Example docker run path
```
docker volume create n8n_data

docker run -d \
  --name n8n \
  -p 5678:5678 \
  -e GENERIC_TIMEZONE="<YOUR_TIMEZONE>" \
  -e TZ="<YOUR_TIMEZONE>" \
  -e N8N_ENFORCE_SETTINGS_FILE_PERMISSIONS=true \
  -v n8n_data:/home/node/.n8n \
  ghcr.io/gonkagate/n8n-nodes-gonkagate:latest
```

For production, pin an exact image tag instead of `latest`. If you prefer Compose, copy the [self-hosted Docker example](https://github.com/GonkaGate/n8n-nodes-gonkagate/tree/main/examples/docker/self-hosted) from the public repository.

## First working request in n8n

After installation, prove the smallest working path first.

1. Open `n8n` and click `Start from scratch`.
2. Add `Manual Trigger`.
3. Click `+` and search for `gonka` or `GonkaGate`.
4. If the picker opens on `AI Nodes`, check `Results in other categories`.
5. Choose the plain `GonkaGate` node for the first validation run.
6. Set `Operation` to `List Models`.
7. Create `GonkaGate API`, paste your API key, and save.
8. Run the node with `Execute step` or `Execute workflow`.

If `List Models` works, switch the same node to `Chat Completion`, choose a model, and send one short message such as `Hello from n8n`.

## Which node should you use?

| Start with… | Use it when… | Why |
|---|---|---|
| GonkaGate | You want the fastest first request, List Models, or easier debugging | Smallest setup surface and direct request/response path |
| GonkaGate Chat Model | You are building AI Agent or other AiLanguageModel workflows | This is the additive AI-model surface for broader n8n AI flows |
Do not start with `GonkaGate Chat Model` for the first validation run. If search also shows `GonkaGate Tool`, skip it for the first check too.

## Common first failures

### The node does not appear after install

- Restart `n8n`.
- If you use queue mode, workers, or webhook runners, make sure the same package version is installed on every runtime process.

### No models appear in the live list

- `/v1/models` can return an empty set.
- Switch the model field to `ID` mode and enter a manual model ID.

### You expected visible streaming in the root node

- The root `GonkaGate` node returns one final JSON response.
- Use `GonkaGate Chat Model` inside a streaming-capable AI workflow if you need visible live streaming.

### The credential fails immediately

- Recreate the credential if it was created before the hidden base URL default was added.
- The expected base URL is `https://api.gonkagate.com/v1`.

## Current boundaries

- Self-hosted-first only, with no `n8n` Cloud promise.
- No blanket compatibility claim across all `n8n` versions.
- Current package scope is `GET /v1/models` plus `POST /v1/chat/completions`.
- No `/v1/responses` support today.

## See also

- [Framework and Tool Guides](/en/docs/guides/community/frameworks-and-integrations-overview) for the broader community hub
- [Claude Code](/en/docs/guides/coding-agents/claude-code), [Cursor](/en/docs/guides/coding-agents/cursor), [Kilo Code](/en/docs/guides/coding-agents/kilo-code), [MiMoCode](/en/docs/guides/coding-agents/mimo-code), [OpenCode](/en/docs/guides/coding-agents/opencode), [OpenClaw](/en/docs/guides/coding-agents/openclaw), and [Hermes Agent](/en/docs/guides/coding-agents/hermes-agent) for agent-owned setup paths
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation and rotation
- [Get models](/en/docs/api/api-reference/models/get-models) for current model IDs when the live picker is empty
- [n8n package repository](https://github.com/GonkaGate/n8n-nodes-gonkagate)