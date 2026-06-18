# Chat Completion Parameters

Tune /v1/chat/completions in GonkaGate: what to change first, what the common fields control, and when to use JSON output or tool calling.

Tune `/v1/chat/completions` one field at a time. Start with `temperature` for output style or `max_tokens` for response size. Move to `response_format`, `tools`, `tool_choice`, or token-level fields only when you need structured output, function calls, or debugging.

## Minimum working example
**Response Example**
```
export GONKAGATE_API_KEY="gp-your-api-key"

curl https://api.gonkagate.com/v1/chat/completions \
  -H "Authorization: Bearer $GONKAGATE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    "messages": [
      {
        "role": "user",
        "content": "Explain idempotency keys in 3 short bullets."
      }
    ],
    "temperature": 0.2,
    "max_tokens": 180
  }'
```

Expected result: the request shape stays the same, but the answer should be steadier and less likely to run long.

- `temperature` makes the answer steadier.
- `max_tokens` caps response size.
- Replace the example `model` with a fresh ID from `GET /v1/models` before rollout.

## Pick the right parameter family

- More stable or more varied text: start with `temperature` or `top_p`. These change sampling behavior.
- Shorter or bounded responses: start with `max_tokens` or `stop`. These control response length and stop conditions.
- Less repetition: start with `frequency_penalty` or `presence_penalty`. These reduce repeated tokens or push the model toward new ones.
- Machine-readable output: start with `response_format`. This changes the response contract, not just the style.
- Function execution from your app: start with `tools` and `tool_choice`. This lets the model request a named function.
- Repeatability or token-level inspection: start with `seed`, `logprobs`, or `top_logprobs`. These help with debugging, evaluation, and reproducibility.

## What the common fields change

### Sampling and repetition

| Field | Type | What it changes | Notes |
|---|---|---|---|
| temperature | number | Randomness and variety | 0-2 |
| top_p | number | Nucleus sampling | 0-1 |
| frequency_penalty | number | Penalizes tokens that repeat often | -2 to 2 |
| presence_penalty | number | Penalizes tokens that already appeared | -2 to 2 |
| stop | string or array | Ends generation on one or more stop sequences | Use when your app already knows the delimiter or boundary. |
Change one sampling family at a time. If you are already adjusting `temperature`, avoid piling on `top_p`, penalties, and `stop` in the same test unless you need them for one very specific contract.

### Response length, repeatability, and debugging

| Field | Type | What it changes | Notes |
|---|---|---|---|
| max_tokens | integer | Upper bound on generated tokens | Set this early if response size matters. Must be at least 1. |
| seed | integer | Best-effort repeatability | Useful for tests and comparisons, but not every model guarantees identical output. |
| logprobs | boolean | Returns token log probabilities | Useful for inspection and evaluation. |
| top_logprobs | integer | Returns the top alternative tokens at each position | 0-20 and only useful with logprobs: true. |

### Response format and tool calling

| Field | Type | What it changes | Notes |
|---|---|---|---|
| response_format | object | Asks the model for JSON or another structured response format | Support depends on the selected model/provider pair. |
| tools | array | Declares functions the model may call | Use the standard OpenAI-compatible tool shape. |
| tool_choice | string or object | Controls whether the model may call a tool and which one | Use required or a specific function when tool use is mandatory. |
GonkaGate stays OpenAI-compatible for `/v1/chat/completions`, but capabilities such as structured outputs and tool calling still depend on the selected model/provider pair.

For the exact request-body schema, validation rules, response payloads, and GonkaGate-specific request fields, use [Create a chat completion](/en/docs/api/api-reference/chat/send-chat-completion-request).

## Common mistakes

- Changing sampling, penalties, output format, and tool settings in the same test, then not knowing which field changed the result.
- Treating `response_format`, `tools`, or `tool_choice` as cosmetic. They change how your application consumes the response.
- Setting `top_logprobs` without `logprobs: true`.
- Assuming structured outputs or tool calling behave the same across every model/provider pair.
- Leaving `max_tokens` unset when response size matters operationally.

## See also

- [Create a chat completion](/en/docs/api/api-reference/chat/send-chat-completion-request) for the full request schema, response payloads, and validation rules.
- [Structured Outputs](/en/docs/guides/features/structured-outputs) for JSON response workflows.
- [Tool Calling](/en/docs/guides/features/tool-calling) for function execution loops.
- [Streaming Responses](/en/docs/api/reference/streaming) for `stream: true` and SSE handling.