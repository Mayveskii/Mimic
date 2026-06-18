# Rate Limit Handling

How to handle GonkaGate 429 responses, current request and token limits, throttling, and retry budgets safely.

Read `error.code` before you retry a `429`. In GonkaGate, `insufficient_quota` means the prepaid USD balance is too low for this request. `rate_limit_exceeded` and `transfer_agent_capacity_reached` are usually temporary. Keep that branch logic in one shared helper so every caller follows the same retry policy.

## Decide whether to retry

| If you get | What it usually means | What to do |
|---|---|---|
| 429 + insufficient_quota | The prepaid USD balance is too low for this request | Stop retrying. Surface balance or top-up state, then retry only after funds are available. |
| 429 + rate_limit_exceeded | Your traffic hit a request limit | Honor Retry-After when present and retry with a small backoff budget. |
| 429 + transfer_agent_capacity_reached | Temporary capacity pressure | Wait, retry carefully, and keep the retry budget small. |
| 429 without a known code | This is usually still temporary throttling or capacity pressure | Treat it as temporary throttling first, but log the full response details if it repeats. |

## Current authenticated model-request limits

For `POST /v1/chat/completions` and other authenticated `/v1/*` model requests, GonkaGate checks multiple rate-limit buckets. The most restrictive exhausted bucket wins.

| Scope | Request and token limits | Burst and concurrency |
|---|---|---|
| Regular API key (gp-...) | 120 RPM and 5,000,000 TPM | 40 request burst and 10 concurrent requests |
| Regular API key + source IP | 120 RPM and 5,000,000 TPM | 40 request burst |
| Owning account | 12,000 RPM and 300,000,000 TPM | 4,000 request burst and 1,000 concurrent requests |
| Source IP | 12,000 RPM and 300,000,000 TPM | 4,000 request burst |
| Distinct regular keys from one source IP | 5,000 keys per hour | Applies when many customer keys exit through the same backend IP |
RPM means requests per minute. TPM means estimated tokens per minute.

These are traffic limits, not spend limits. A per-key USD `limit` configured through API key management caps spending for that regular key, while the account prepaid USD balance remains shared across the account.

## Use one shared retry helper
 Use one shared retry helper
```
const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export async function requestWithRateLimitHandling(
  makeRequest: () => Promise<Response>,
  maxRetries = 3
): Promise<Response> {
  for (let attempt = 0; attempt <= maxRetries; attempt += 1) {
    const response = await makeRequest();

    if (response.status !== 429) {
      return response;
    }

    const body = await response
      .clone()
      .json()
      .catch(() => null);
    const errorCode = body?.error?.code;

    if (errorCode === "insufficient_quota") {
      throw new Error("insufficient_quota");
    }

    if (attempt === maxRetries) {
      throw new Error("retry_budget_exhausted");
    }

    const retryAfterSeconds = Number(response.headers.get("Retry-After") ?? "0");
    const waitMs =
      retryAfterSeconds > 0 ? retryAfterSeconds * 1000 : Math.min(1000 * 2 ** attempt, 8000);

    await sleep(waitMs);
  }

  throw new Error("retry_budget_exhausted");
}
```

This baseline does four things: branch on `error.code`, stop on `insufficient_quota`, honor `Retry-After`, and cap retries.

## Read these fields first

- HTTP status `429`
- `error.code` in the JSON body
- `Retry-After` when the server tells you exactly how long to wait
- `x-ratelimit-*` headers when your client exposes them, so you can log the current limit, remaining allowance, and reset window
- `x-request-id` for repeated failures or support escalation

Treat `error.message` as human-readable context only. Do not build retry logic from the message text.

## Common mistakes

- Treating every `429` as retryable throttling. In GonkaGate, `insufficient_quota` is a billing state, not a backoff case.
- Ignoring `Retry-After` when it is present. That usually creates synchronized retries and more throttling.
- Hiding `insufficient_quota` behind automatic retries. Stop and show a billing or top-up state instead.
- Assuming that many generated `gp-...` keys each get a separate account-level quota. Per-key buckets are separate, but keys under one GonkaGate account still share that account’s aggregate bucket.
- Letting workers, cron jobs, or batch traffic retry forever. Keep the retry budget small and make interactive traffic the priority.

## See also

- [Management API Keys](/en/docs/authentication/management-api-keys) for creating regular `gp-...` keys programmatically and setting per-key USD spend limits.
- [GonkaGate API Error Handling](/en/docs/api/reference/error-handling) for the same retry-or-stop logic across `401`, `403`, `5xx`, and other non-`429` failures.
- [Pricing](/en/pricing) for prepaid USD balance rules behind `insufficient_quota`.
- [Create a chat completion](/en/docs/api/api-reference/chat/send-chat-completion-request) for the exact `POST /v1/chat/completions` request and response contract.