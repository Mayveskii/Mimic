# Tool Calling (Function Calling)

Let the model request tools or functions from your application through GonkaGate chat completions.

Use tool calling (also known as function calling) when the model should request work from your backend or another external system. The model does not run the tool itself. It returns `tool_calls`; your application validates the arguments, runs the matching handler, appends one `role: "tool"` message per `tool_call_id`, and sends the updated conversation back until the model returns a normal message.

Keep your own iteration limit, validate every argument before execution, and re-check authorization inside each handler. If you only need machine-readable JSON, use [Structured Outputs](./structured-outputs). If GonkaGate should run the behavior for you, use [Chat Completions Plugins](./plugins/overview).

## Before you start

- A working `chat.completions` request against `https://api.gonkagate.com/v1`.
- A model/provider pair that supports tool calling for your use case.
- One clear server-side handler for each tool your application exposes.
- A permission check inside each handler for the current user or session. A model-emitted tool call is a request, not proof that the action is allowed.
- Argument validation before execution, plus `JSON.stringify(...)` when you send structured tool results back.

## Run the loop

1. Send `messages`, `tools`, and optional `tool_choice` to `/v1/chat/completions`.
2. If the model returns `tool_calls`, validate the arguments and run the matching handler in your application.
3. Append the assistant message with `tool_calls`, append one `role: "tool"` message per `tool_call_id`, and send the conversation back, usually with the same `tools` again if later turns may need another tool.

Repeat until the model returns a message without `tool_calls`, and stop when you hit your own iteration limit.

GonkaGate forwards `tools` and `tool_choice` through the same OpenAI-compatible `/v1/chat/completions` API. Your application owns argument validation, execution, timeouts, retries, the follow-up request, and authorization. Tool calls are still untrusted model output.

## Send the first request

The request shape is the same OpenAI-compatible `chat.completions` payload plus `tools` and optional `tool_choice`.
 request.json
```
{
  "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
  "messages": [
    {
      "role": "user",
      "content": "Check whether api-gateway is healthy in prod and summarize the result."
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_deployment_status",
        "description": "Return the current deployment status for a service",
        "parameters": {
          "type": "object",
          "properties": {
            "service": { "type": "string" },
            "environment": {
              "type": "string",
              "enum": ["prod", "staging"]
            }
          },
          "required": ["service", "environment"]
        }
      }
    }
  ],
  "tool_choice": "auto"
}
```

## Append the tool result

If the model wants your application to do work, it can return a tool call like this:
 assistant-message.json
```
{
  "role": "assistant",
  "content": null,
  "tool_calls": [
    {
      "id": "call_abc123",
      "type": "function",
      "function": {
        "name": "get_deployment_status",
        "arguments": "{\"service\":\"api-gateway\",\"environment\":\"prod\"}"
      }
    }
  ]
}
```

After you run the handler locally, append one tool message for that `tool_call_id`:
 tool-message.json
```
{
  "role": "tool",
  "tool_call_id": "call_abc123",
  "content": "{\"service\":\"api-gateway\",\"environment\":\"prod\",\"status\":\"healthy\"}"
}
```

Then send the updated conversation back to the model. In most integrations you also pass the same `tools` array again so the model can request another tool on the next turn.

Expected result: after one or more tool rounds, the final assistant message contains normal `content` and no `tool_calls`.

## End-to-end example

This is the minimal safe pattern: validate arguments, reject unknown tools, turn parse and handler failures into explicit tool errors, resend `tools` on follow-up turns, and stop after your own loop limit.
 End-to-end example
```
import OpenAI from "openai";

type Environment = "prod" | "staging";

const client = new OpenAI({
  baseURL: "https://api.gonkagate.com/v1",
  apiKey: "gp-your-api-key"
});

const model = "qwen/qwen3-235b-a22b-instruct-2507-fp8";

const tools: OpenAI.Chat.Completions.ChatCompletionTool[] = [
  {
    type: "function",
    function: {
      name: "get_deployment_status",
      description: "Return the current deployment status for a service",
      parameters: {
        type: "object",
        properties: {
          service: { type: "string" },
          environment: {
            type: "string",
            enum: ["prod", "staging"]
          }
        },
        required: ["service", "environment"]
      }
    }
  }
];

async function getDeploymentStatus(service: string, environment: Environment) {
  return {
    service,
    environment,
    status: "healthy",
    updated_at: "2026-03-14T09:30:00Z"
  };
}

type ToolCall = {
  id: string;
  function: {
    name: string;
    arguments: string | null;
  };
};

function parseArgs(raw: string): { service: string; environment: Environment } {
  const parsed = JSON.parse(raw || "{}") as {
    service?: unknown;
    environment?: unknown;
  };

  if (typeof parsed.service !== "string" || parsed.service.trim() === "") {
    throw new Error("service must be a non-empty string");
  }

  if (parsed.environment !== "prod" && parsed.environment !== "staging") {
    throw new Error("environment must be prod or staging");
  }

  return {
    service: parsed.service,
    environment: parsed.environment
  };
}

const messages: OpenAI.Chat.Completions.ChatCompletionMessageParam[] = [
  {
    role: "user",
    content: "Check whether api-gateway is healthy in prod and summarize the result."
  }
];

async function executeToolCall(
  call: ToolCall
): Promise<OpenAI.Chat.Completions.ChatCompletionToolMessageParam> {
  if (call.function.name !== "get_deployment_status") {
    return {
      role: "tool",
      tool_call_id: call.id,
      content: JSON.stringify({
        ok: false,
        error: `unknown tool: ${call.function.name}`
      })
    };
  }

  try {
    const args = parseArgs(call.function.arguments || "{}");
    const result = await getDeploymentStatus(args.service, args.environment);

    return {
      role: "tool",
      tool_call_id: call.id,
      content: JSON.stringify({
        ok: true,
        result
      })
    };
  } catch (error) {
    return {
      role: "tool",
      tool_call_id: call.id,
      content: JSON.stringify({
        ok: false,
        error: error instanceof Error ? error.message : "tool_execution_failed"
      })
    };
  }
}

const maxToolRounds = 5;
let finalText: string | null = null;

for (let round = 0; round < maxToolRounds; round++) {
  const response = await client.chat.completions.create({
    model,
    messages,
    tools,
    tool_choice: "auto"
  });

  const assistantMessage = response.choices[0]?.message;
  const toolCalls = assistantMessage?.tool_calls ?? [];

  if (!assistantMessage) {
    throw new Error("Model returned no message");
  }

  messages.push(assistantMessage);

  if (!toolCalls.length) {
    finalText = assistantMessage.content ?? "";
    break;
  }

  const toolMessages = await Promise.all(toolCalls.map(executeToolCall));

  messages.push(...toolMessages);
}

if (finalText === null) {
  throw new Error(`Tool loop exceeded ${maxToolRounds} rounds`);
}

console.log(finalText);
```

## Common failures

- No `tool_calls` appear: the model may answer directly. Tighten the prompt or set `tool_choice` to `required` or a specific function when the action is mandatory.
- `function.arguments` parse but are still wrong: treat them as untrusted JSON. Validate enums, required fields, and empty strings before calling your backend.
- `function.arguments` do not parse, the handler throws, or the tool times out: return an explicit tool error payload or abort the loop cleanly, but do not pretend the tool succeeded.
- The model seems to ignore the tool result: append the assistant message with `tool_calls` before you append `role: "tool"` messages, then make the follow-up request.
- The tool call targets a privileged action: do not treat the model’s request as authorization. Re-check the current user or session before the handler runs.
- Multiple tool calls arrive in one turn: return one `role: "tool"` message per `tool_call_id` and keep the IDs unchanged.
- You forget to send `tools` again on the follow-up request: if later turns may need more tool calls, resend the same `tools` array on each round.
- Streamed tool calls arrive in deltas: rebuild the full `arguments` string before you parse JSON or execute the handler.
- Tool results fail schema checks in your client: `content` in the tool message is a string. If the handler returns JSON, send it back with `JSON.stringify(...)`.

## See also

- [Chat Completion Parameters](/en/docs/api/reference/parameters) for the exact `tools` and `tool_choice` field contract.
- [Structured Outputs](./structured-outputs) for schema-shaped JSON without app-executed actions.
- [Chat Completions Plugins](./plugins/overview) for GonkaGate-managed behavior such as built-in web search or response healing.