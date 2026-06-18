# Chat Completion Presets

Reuse saved defaults in /v1/chat/completions with presets.

Use chat completion presets when the same defaults should apply across many `/v1/chat/completions` requests. Start with explicit `model` plus `preset`. The request stays readable and OpenAI-compatible, while the preset fills only missing supported defaults.

## Minimum working request
**Request Example**
```
import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "https://api.gonkagate.com/v1",
  apiKey: "gp-your-api-key"
});

const response = await client.chat.completions.create({
  model: "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  preset: "support-agent",
  messages: [{ role: "user", content: "Draft a concise reply to this support ticket." }],
  temperature: 0.3
});

console.log(response.choices[0]?.message?.content);
```

Expected behavior: request-level fields still win. In this example, `temperature: 0.3` overrides the preset value, while other missing supported fields can still come from the preset.

## Choose the attach pattern

- `model: "<model>"` plus `preset: "<slug>"`: start here when the request should stay explicit and easy to debug. GonkaGate keeps the request model and fills only missing supported defaults from the preset.
- `model: "@preset/<slug>"`: use this when the preset should choose model order too. GonkaGate uses the preset `models` list, system prompt, `params`, and `reasoning`.
- `model: "<model>@preset/<slug>"`: use this when one exact model is fixed, but the preset should still supply the other defaults. GonkaGate keeps the explicit request model and still applies the preset system prompt and supported defaults.

## What a preset can store

Use presets for stable shared defaults. Keep request-level fields for values that change from call to call.

- `systemPrompt`
- `params`
- `reasoning`
- `models`

**Request Example**
```
{
  "name": "Support Agent",
  "slug": "support-agent",
  "description": "Preset for support replies",
  "systemPrompt": "You are a concise support assistant.",
  "models": ["qwen/qwen3-235b-a22b-instruct-2507-fp8"],
  "params": {
    "temperature": 0.2,
    "top_p": 0.9
  },
  "reasoning": {
    "enabled": true,
    "effort": "high"
  }
}
```

Validation limits that affect integration:

- `models` supports up to 10 entries.
- `params` accepts only `temperature`, `top_p`, `top_k`, `frequency_penalty`, `presence_penalty`, `repetition_penalty`, `max_tokens`, and `seed`.
- `reasoning.enabled` is required when the `reasoning` block is present.
- `reasoning.effort` and `reasoning.max_tokens` cannot be used together.

## How preset merge and precedence work

1. GonkaGate starts with the request fields.
2. Missing supported fields are filled from preset `params` and `reasoning`.
3. The `preset` field is removed before the request is forwarded upstream.
4. Preset `systemPrompt` is applied as a system message:
- if the request has no system message, GonkaGate inserts the preset system prompt first
- if all existing system messages use string content, GonkaGate merges them into one system message, with the preset text first and blank lines between parts
- if any existing system message uses non-string content, GonkaGate inserts the preset system prompt as a separate system message before the others

For supported preset defaults, precedence is `request > preset > default`. Presets change the input before it goes upstream, but they do not change the OpenAI-compatible response shape.

## Models and fallback

- `model: "@preset/<slug>"` requires at least one saved model in the preset `models` list.
- If a preset has `models`, GonkaGate builds an ordered candidate list from that array.
- Retry-based fallback runs only for retry-eligible errors such as transient provider or model-availability problems.
- If the request uses `<model>@preset/<slug>` or explicit `model` plus `preset`, fallback stays on that model and does not use the preset model list.
- For streaming requests, fallback can happen only before the first chunk is sent.

## Common errors and limits

- Disabled presets cannot be used in `/v1/chat/completions`; those requests fail with `preset_disabled`.
- `preset_not_found`: the slug does not point to an available preset.
- `preset_missing_model`: the preset has no usable model list and the request did not override the model.
- `preset_ambiguous`: `model` and `preset` point to different preset slugs.
- `preset_invalid`: the `@preset/` reference is malformed, for example `{"model":"@preset/"}`.
- `preset_invalid_slug`: the slug format is invalid or reserved, for example `{"model":"@preset/Ad"}`.

## Create and manage presets in Dashboard

Presets are per-user and are managed in Dashboard. Create, edit, enable, disable, version, and roll them back there. Runtime usage still happens through normal `/v1/chat/completions` requests.

Slug rules:

- 3 to 64 characters
- lowercase `a-z`, digits, and `-` only
- no spaces, uppercase letters, leading or trailing hyphens, or repeated hyphen groups
- if `slug` is omitted, it is generated from `name`

Rollback creates a new current version from a historical snapshot. It does not just point the preset back to an older record.

## See also

- [Chat Completions API reference](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact request schema.
- [Chat Completion Parameters](/en/docs/api/reference/parameters) for field-level parameter behavior.
- [Structured Outputs](./structured-outputs) when the response must follow a JSON contract.
- [Tool Calling](./tool-calling) when the model should call your application tools.
- [Choose a Plugin](./plugins/overview) for request-level runtime extensions.