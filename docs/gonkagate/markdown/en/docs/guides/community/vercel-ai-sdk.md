# Vercel AI SDK Setup

Connect an existing Vercel AI SDK app to GonkaGate.

Connect an existing Vercel AI SDK app to GonkaGate by creating one OpenAI-compatible provider, using a current model ID, and returning `result.toTextStreamResponse()` from your server route. Keep `GONKAGATE_API_KEY` in the same server runtime that handles the route.

## Install the provider package
 Installation
```
npm install @ai-sdk/openai-compatible
```

## Create one shared provider

Create one provider and reuse it across your routes:
 Create one shared provider
```
import { createOpenAICompatible } from "@ai-sdk/openai-compatible";

const apiKey = process.env.GONKAGATE_API_KEY;

if (!apiKey) {
  throw new Error("Set GONKAGATE_API_KEY");
}

export const gonkagate = createOpenAICompatible({
  name: "gonkagate",
  apiKey,
  baseURL: "https://api.gonkagate.com/v1"
});
```

## Use it in a server route

Use the provider in a server route that returns streamed text:
 Use it in a server route
```
import { streamText } from "ai";

import { gonkagate } from "@/lib/gonkagate";

export async function POST(request: Request) {
  const { prompt } = (await request.json()) as { prompt: string };

  const result = streamText({
    model: gonkagate("<model-id-from-get-v1-models>"),
    prompt
  });

  return result.toTextStreamResponse();
}
```

Replace `<model-id-from-get-v1-models>` with a current value from [GET /v1/models](/en/docs/api/api-reference/models/get-models).

## Verify the stream

Send one short request through the route:
**Request Example**
```
curl http://localhost:3000/api/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt":"Return exactly: AI SDK connected"}'
```

Expected result: the route returns streamed text that includes `AI SDK connected`.

## Common failures

| Response or symptom | What it usually means | What to do |
|---|---|---|
| GONKAGATE_API_KEY is missing | The variable is not loaded in the server runtime that handles the route | Load it in the same runtime where the route runs |
| 401 invalid_api_key | The key value or key state is wrong | Recheck Authentication and API Keys |
| 404 model_not_found | The model ID is stale or unsupported | Refresh it from GET /v1/models |
| 429 insufficient_quota | The prepaid USD balance is too low for the request | Top up before retrying |
| 429 rate_limit_exceeded | You hit a runtime limit | Honor Retry-After and add bounded backoff |
| The route responds but the client still handles the stream incorrectly | The route and caller no longer agree on a text-stream response | Keep result.toTextStreamResponse() on the route and make sure the caller expects streamed text |

## See also

- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you need to move more of an existing OpenAI-compatible app
- [TypeScript SDK for GonkaGate](/en/docs/sdk/typescript) if you do not want the AI SDK provider layer
- [Streaming reference](/en/docs/api/reference/streaming) for exact SSE behavior, `usage` fields, and streaming edge cases