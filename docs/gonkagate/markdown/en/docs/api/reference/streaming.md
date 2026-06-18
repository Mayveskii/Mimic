# Streaming Responses

Enable streaming on POST /v1/chat/completions and parse SSE chunks from GonkaGate.

Set `stream: true` on OpenAI-compatible `POST /v1/chat/completions` to receive Server-Sent Events (SSE) from GonkaGate as tokens are generated. Read each `data:` event incrementally, ignore keep-alive comments, and treat the response as complete only after `[DONE]`.

Start from a working non-streaming request. Use streaming for chat UIs, long responses, or operator flows that need partial output. If you only need the final answer, the JSON response path is simpler to ship and debug.

## Minimum working example
**Request Example**
```
export GONKAGATE_API_KEY="gp-your-api-key"

curl -N https://api.gonkagate.com/v1/chat/completions \
  -H "Authorization: Bearer $GONKAGATE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    "messages": [
      { "role": "user", "content": "Explain SSE in one sentence." }
    ],
    "stream": true
  }'
```

Use `-N` so `curl` prints the stream as it arrives. Replace the model ID with one available to your account. You also need a valid API key and enough prepaid USD balance for the request.

## How the SSE stream is structured

A successful streaming response is a sequence of SSE events, not one JSON document:
**Response Example**
```
data: {"id":"chatcmpl-abc","object":"chat.completion.chunk","created":1735500000,"model":"qwen/qwen3-32b-fp8","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc","object":"chat.completion.chunk","created":1735500000,"model":"qwen/qwen3-32b-fp8","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-abc","object":"chat.completion.chunk","created":1735500000,"model":"qwen/qwen3-32b-fp8","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}

data: [DONE]
```

Read it in this order:

- The first chunk can establish the assistant role and contain no text yet.
- Text usually arrives in later `choices[0].delta.content` chunks.
- The last text-bearing chunk can carry `finish_reason`.
- Ignore keep-alive comments such as `: keep-alive` and other non-`data:` lines.
- The stream is complete only after `[DONE]`.

## Minimal parser example
 TypeScript (fetch)
```
const response = await fetch("https://api.gonkagate.com/v1/chat/completions", {
  method: "POST",
  headers: {
    Authorization: `Bearer ${process.env.GONKAGATE_API_KEY}`,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    model: "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    messages: [{ role: "user", content: "Explain SSE in one sentence." }],
    stream: true
  })
});

if (!response.ok) {
  throw new Error(`Streaming request failed with ${response.status}`);
}

const reader = response.body?.getReader();
if (!reader) {
  throw new Error("Response body is not readable");
}

const decoder = new TextDecoder();
let buffer = "";
let streamDone = false;
let finalUsage: Record<string, unknown> | null = null;

while (!streamDone) {
  const { done, value } = await reader.read();

  if (done) {
    throw new Error("Stream ended before [DONE]");
  }

  buffer += decoder.decode(value, { stream: true });

  while (true) {
    const newlineIndex = buffer.indexOf("\n");
    if (newlineIndex === -1) {
      break;
    }

    const line = buffer.slice(0, newlineIndex).trim();
    buffer = buffer.slice(newlineIndex + 1);

    if (!line.startsWith("data:")) {
      continue;
    }

    const payload = line.slice(5).trim();

    if (payload === "[DONE]") {
      streamDone = true;
      break;
    }

    const chunk = JSON.parse(payload) as {
      error?: { message?: string; type?: string; code?: string };
      choices?: Array<{ delta?: { content?: string } }>;
      usage?: Record<string, unknown>;
    };

    if (chunk.error) {
      throw new Error(
        `Streaming error (${chunk.error.code ?? chunk.error.type ?? "unknown"}): ${chunk.error.message ?? "Unknown error"}`
      );
    }

    const content = chunk.choices?.[0]?.delta?.content;
    if (content) {
      process.stdout.write(content);
    }

    if (chunk.usage) {
      finalUsage = chunk.usage;
    }
  }
}

console.log("\nFinal usage:", finalUsage);
```

This keeps the parser simple: ignore non-`data:` lines, parse each complete SSE payload, stream `delta.content` as it arrives, stop on an SSE error event, fail on premature EOF, and wait for the final usage chunk before treating token usage as final.

## Common mistakes and failure cases

- A client that waits for one full JSON body will look “stuck” even when the stream is healthy. Use an SSE-capable client or read the response incrementally.
- Do not mark the response complete on the first token. Finish only after the stream reaches `[DONE]`.
- Do not finalize token usage before the final usage chunk arrives. Use the dashboard for completed spend.
- If an error happens after streaming starts, do not expect a non-`200` HTTP status. Handle `data: {"error": ...}` as a terminal SSE event.
- Do not build parser logic around GonkaGate-specific prelude events for public `privacy-sanitization`. Keep the parser focused on normal SSE `data:` payloads, keep-alive comments, and `[DONE]`.
- Mid-stream disconnects are incomplete responses. Keep retry, fallback, and user-visible failure handling outside the parser itself.

## See also

- [Create a chat completion](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact `POST /v1/chat/completions` request fields, SSE schema, and endpoint contract.
- [Quickstart](/en/docs/quickstart) if you still need the first successful non-streaming request.
- [API Error Handling](/en/docs/api/reference/error-handling) for retry vs stop decisions after stream failures.