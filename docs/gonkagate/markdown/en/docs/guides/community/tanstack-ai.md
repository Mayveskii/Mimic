# TanStack AI Setup

Connect TanStack AI chat flows to GonkaGate.

Connect TanStack AI chat flows to GonkaGate with one server route, one React client, and the OpenAI adapter. Keep `GONKAGATE_API_KEY` on the server and use a current model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models). This guide covers the `useChat` plus SSE path. If you work directly with AI SDK primitives such as `streamText`, use [Vercel AI SDK Setup for GonkaGate](/en/docs/guides/community/vercel-ai-sdk) instead.

## Install the packages
 Installation
```
npm install @tanstack/ai @tanstack/ai-react @tanstack/ai-openai
```

## Set `GONKAGATE_API_KEY`
 Set GONKAGATE_API_KEY
```
GONKAGATE_API_KEY=your_api_key_here
```

Keep the key on the server. Do not expose it in browser code.

## Create one server route and one React client

The smallest useful setup is one server route plus one React client that both use the same GonkaGate chat path.

### Server route
 Server route
```
import { chat, toServerSentEventsResponse } from "@tanstack/ai";
import { createOpenaiChat } from "@tanstack/ai-openai";

const apiKey = process.env.GONKAGATE_API_KEY;

if (!apiKey) {
  throw new Error("Set GONKAGATE_API_KEY");
}

const gonkagate = createOpenaiChat(apiKey, {
  baseURL: "https://api.gonkagate.com/v1"
});

export async function POST(request: Request) {
  const { messages } = await request.json();

  const stream = chat({
    adapter: gonkagate("qwen/qwen3-235b-a22b-instruct-2507-fp8"),
    messages
  });

  return toServerSentEventsResponse(stream);
}
```

### React client
 React client
```
"use client";

import { useState } from "react";

import { fetchServerSentEvents, useChat } from "@tanstack/ai-react";

export function Chat() {
  const [input, setInput] = useState("");
  const { messages, sendMessage, isLoading, error } = useChat({
    connection: fetchServerSentEvents("/api/chat")
  });

  return (
    <div>
      {messages.map((message) => (
        <div key={message.id}>
          {message.parts.map((part, index) =>
            part.type === "text" ? <p key={index}>{part.content}</p> : null
          )}
        </div>
      ))}

      {error ? <p>{error.message}</p> : null}

      <form
        onSubmit={(event) => {
          event.preventDefault();
          if (!input.trim() || isLoading) return;
          void sendMessage(input);
          setInput("");
        }}
      >
        <input value={input} onChange={(event) => setInput(event.target.value)} />
        <button type="submit" disabled={isLoading}>
          Send
        </button>
      </form>
    </div>
  );
}
```

Replace the example model with a current model ID from [GET /v1/models](/en/docs/api/api-reference/models/get-models).

## Verify the chat path

Send `Return exactly: TanStack AI connected` through the chat UI. Success means streamed assistant text reaches the client through `useChat`.

## Keep the SSE pair together

- Keep `toServerSentEventsResponse(stream)` on the server and `fetchServerSentEvents("/api/chat")` on the client.
- Keep `GONKAGATE_API_KEY` in the server runtime that handles the route.

## Common failures

| Response or symptom | What it usually means | What to do |
|---|---|---|
| 401 invalid_api_key | The server route cannot read GONKAGATE_API_KEY, or the key is invalid | Recheck Authentication and API Keys |
| 404 model_not_found | The model ID is stale or unsupported | Refresh it from GET /v1/models or the Model Selection Guide |
| 429 insufficient_quota | The prepaid USD balance is too low for the request | Top up before retrying |
| 429 rate_limit_exceeded | You hit a runtime limit | Honor Retry-After and add bounded backoff |
| The route responds but the UI does not stream | The server and client are no longer using the same SSE path | Return toServerSentEventsResponse(stream) from the route and keep fetchServerSentEvents("/api/chat") on the client |

## See also

- [OpenAI to GonkaGate Migration Guide](/en/docs/guides/overview/migration) if you are switching more than one OpenAI-compatible chat surface
- [Vercel AI SDK Setup for GonkaGate](/en/docs/guides/community/vercel-ai-sdk) if your app works directly with AI SDK primitives such as `streamText`
- [TanStack AI OpenAI adapter docs](https://tanstack.com/ai/latest/docs/adapters/openai) for adapter-specific options
- [TanStack AI React chat example](https://github.com/TanStack/ai/tree/main/examples/ts-react-chat) for a larger end-to-end example