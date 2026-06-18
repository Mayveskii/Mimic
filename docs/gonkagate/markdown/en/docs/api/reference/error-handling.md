# API Error Handling

Retry, stop, or fix the request after a GonkaGate API failure.

Decide whether to retry, stop, or fix the request by reading the HTTP status first and `error.code` second. Keep `x-request-id` on every failed request so repeated failures are easy to trace and escalate.

## Retry, stop, or fix the request

| If you see | What it usually means | What to do |
|---|---|---|
| 401 or 403 | Auth or account-state problem | Stop retrying. Fix the API key, email, account state, or access first. |
| 429 + insufficient_quota | The prepaid USD balance is too low for this request | Do not back off and retry. Show balance or top-up state. Retry only after funds are available. |
| Other 429 | Throttling or temporary capacity pressure | Respect Retry-After when present. Retry with a small backoff budget. |
| 5xx, timeout, or connection reset | Transient platform or upstream failure | Retry with a small budget. If the same failure repeats, escalate with x-request-id and request context. |
| Other 4xx | The request shape is wrong, the selected model is unavailable, or the input is unsupported | Fix the request. Do not retry blindly. |

## Read HTTP status first

For non-success responses, expect an error object in the JSON body:
**Error Example**
```
type GonkaErrorResponse = {
  error: {
    message: string;
    type?: string;
    code?: string;
  };
};
```

Branch on HTTP status first, then `error.code`. Use `error.message` for logs or UI only. Do not build retry logic from message text. If the body is missing or invalid JSON, keep the same status-based branch and log `x-request-id`.

## Implement one shared decision helper
**Request Example**
```
type GonkaErrorResponse = {
  error?: {
    message?: string;
    type?: string;
    code?: string;
  };
};

function decideGonkaAction(status: number, code?: string) {
  if (status === 401 || status === 403) return "fix-auth-or-account";
  if (status === 429 && code === "insufficient_quota") return "show-billing-state";
  if (status === 429) return "retry-with-backoff";
  if (status >= 500) return "retry-then-escalate";
  return "fix-request";
}

async function sendRequest() {
  const response = await fetch("https://api.gonkagate.com/v1/chat/completions", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${process.env.GONKAGATE_API_KEY}`,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      model: "qwen/qwen3-235b-a22b-instruct-2507-fp8",
      messages: [{ role: "user", content: "Hello, GonkaGate!" }]
    })
  });

  const requestId = response.headers.get("x-request-id") ?? "unknown";
  const payload = (await response.json().catch(() => ({}))) as GonkaErrorResponse;

  if (response.ok) {
    console.log("success");
    return;
  }

  const action = decideGonkaAction(response.status, payload.error?.code);

  console.error({
    action,
    requestId,
    status: response.status,
    type: payload.error?.type,
    code: payload.error?.code,
    message: payload.error?.message
  });
}

sendRequest().catch(console.error);
```

Use one shared helper so every caller makes the same retry-or-stop decision for the same failure class.

## Keep this packet for support

Keep this packet when a request fails repeatedly or needs support:

- `x-request-id`
- HTTP status, `error.type`, `error.code`, and `error.message`
- model ID, latency, and retry count
- sanitized request context
- balance or throttling context for `429` failures

If repeated `5xx`, timeouts, or connection resets continue after the retry budget, escalate with this packet instead of increasing retries forever.

## Common mistakes

- Retrying every `429`. `insufficient_quota` is a billing state, not retryable throttling.
- Matching on `error.message` text instead of HTTP status and `error.code`.
- Hiding `401`, `403`, or balance problems behind background retries.
- Letting workers, cron jobs, or browser tabs retry forever.

## See also

- [GonkaGate Rate Limit Handling](/en/docs/api/reference/rate-limits) for exact `429` backoff, `Retry-After`, and traffic-shaping rules.
- [Authentication and API Keys](/en/docs/authentication/api-keys) for Bearer header checks, key lifecycle, and account-state failures.
- [Pricing](/en/pricing) for prepaid USD balance and billing rules behind `insufficient_quota`.
- [Create a chat completion](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact `POST /v1/chat/completions` contract and endpoint-level examples.