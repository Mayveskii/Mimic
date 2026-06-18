# Python SDK

Send chat.completions requests to GonkaGate from Python with the official OpenAI SDK.

Send one `chat.completions` request from Python with the official OpenAI SDK, then choose `OpenAI` or `AsyncOpenAI` based on your concurrency model.

## Minimum working example

Install `openai` and export your API key:
 Installation
```
pip install openai
export GONKAGATE_API_KEY="gp-your-api-key"
```

Send one `chat.completions` request with the standard OpenAI Python client:
**Request Example**
```
import os
from openai import OpenAI

client = OpenAI(
    base_url="https://api.gonkagate.com/v1",
    api_key=os.environ["GONKAGATE_API_KEY"],
)

response = client.chat.completions.create(
    model="qwen/qwen3-235b-a22b-instruct-2507-fp8",
    messages=[{"role": "user", "content": "Hello from Python"}],
)

print(response.choices[0].message.content)
```

Expected result: you get text back in `response.choices[0].message.content`.

Use a current model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models). The example value above is only illustrative.

## What you need before you run it

- Python 3.9+ with package install access
- A GonkaGate API key that starts with `gp-` in a server-side environment variable or secret store
- Enough available balance for the request
- A current GonkaGate model ID

## Choose the client

Use `OpenAI` for scripts, cron jobs, and simple backends that make one request at a time.

Use `AsyncOpenAI` for APIs, workers, and services that send concurrent requests.
 Choose the client
```
import asyncio
import os
from openai import AsyncOpenAI

client = AsyncOpenAI(
    base_url="https://api.gonkagate.com/v1",
    api_key=os.environ["GONKAGATE_API_KEY"],
)

async def main():
    response = await client.chat.completions.create(
        model="qwen/qwen3-235b-a22b-instruct-2507-fp8",
        messages=[{"role": "user", "content": "Hello from async Python"}],
    )

    print(response.choices[0].message.content)

asyncio.run(main())
```

## Python-specific notes

- Reuse one configured client per process or worker instead of rebuilding it on every request.
- Keep `gp-...` keys on the server side, not in browsers or public notebooks.
- Set `timeout` and `max_retries` explicitly on the client before production traffic.

## Common errors and limits

- `401 invalid_api_key` usually means the key is missing, malformed, or belongs to an unavailable account state.
- `404 model_not_found` means the model ID is not currently available on GonkaGate.
- `429 insufficient_quota` usually means the available prepaid USD balance is too low for the request.

## See also

- [OpenAI SDK compatibility](/en/docs/sdk/openai) for the shared GonkaGate rules that stay the same across runtimes.
- [Authentication and API Keys](/en/docs/authentication/api-keys) for key creation, storage, and rotation.
- [API Reference Overview](/en/docs/api/reference/overview) when your Python integration needs exact request fields, streaming behavior, or retry and failure semantics.
- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you are switching an existing OpenAI-compatible Python app.