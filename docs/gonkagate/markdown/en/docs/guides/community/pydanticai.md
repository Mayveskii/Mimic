# PydanticAI Setup

Set up PydanticAI with GonkaGate.

Keep the rest of your PydanticAI app untouched and swap only the OpenAI-compatible provider values. Use the minimum path below to confirm one request works before you validate tools, output models, graphs, or multi-agent flows.

## Quick setup

If you still need the OpenAI provider extras in PydanticAI:
 Quick setup
```
pip install 'pydantic-ai-slim[openai]'
```

Then point `OpenAIProvider(...)` at GonkaGate and keep the rest of your PydanticAI app on the same wrapper path.
**Python Example**
```
import os

from pydantic_ai import Agent
from pydantic_ai.models.openai import OpenAIChatModel
from pydantic_ai.providers.openai import OpenAIProvider

api_key = os.getenv("GONKAGATE_API_KEY")

if not api_key:
    raise RuntimeError("Set GONKAGATE_API_KEY before running this example")

provider = OpenAIProvider(
    base_url="https://api.gonkagate.com/v1",
    api_key=api_key,
)

model = OpenAIChatModel(
    "<model-id-from-get-v1-models>",
    provider=provider,
)

agent = Agent(model)
result = agent.run_sync("Return exactly: PydanticAI connected")

print(result.output)
```

Replace `<model-id-from-get-v1-models>` with a current value from [GET /v1/models](/en/docs/api/api-reference/models/get-models).

## When this page is the right fit

- Your Python app already uses PydanticAI.
- The app talks through `OpenAIProvider(...)` and `OpenAIChatModel(...)`.
- You only need the GonkaGate connection values, not a full migration or a raw SDK rewrite.

## What stays the same

Keep your existing `Agent(...)` definition and application flow around this provider path. After the base connection works, validate any tools, output models, graphs, or multi-agent flows separately instead of treating this page as the owner for those features.

## Common failures

- `GONKAGATE_API_KEY` is missing: export the variable in the same runtime that starts the agent.
- `401 invalid_api_key`: check key loading and key state in [Authentication and API Keys](/en/docs/authentication/api-keys).
- `404 model_not_found`: switch to a live model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models).
- `429 insufficient_quota`: the prepaid USD balance is too low for this request.
- `429 rate_limit_exceeded`: treat this as throttling. Respect `Retry-After` and add bounded backoff.

## Next step

If your task is bigger than this provider swap, use the [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration).

## Official docs

- [PydanticAI OpenAI models docs](https://ai.pydantic.dev/models/openai/)
- [PydanticAI OpenAI models API reference](https://ai.pydantic.dev/api/models/openai/)