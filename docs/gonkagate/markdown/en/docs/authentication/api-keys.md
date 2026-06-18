# Authentication and API Keys

Authenticate requests with a GonkaGate API key.

Use `GET /v1/models` as a first GonkaGate auth check before you debug request bodies. A `200 OK` response means the Bearer header was accepted, the key passed the current access checks, and the request reached the model catalog.

## Verify the key with one request

1. If you do not have a key yet, sign in and open [Dashboard → API Keys](/en/dashboard/keys) to create one.
2. Save the full secret when GonkaGate shows it.
3. Send it in the `Authorization` header:

**Request Example**
```
export GONKAGATE_API_KEY="gp-your-api-key"

curl https://api.gonkagate.com/v1/models \
  -H "Authorization: Bearer $GONKAGATE_API_KEY"
```

Expected result: `200 OK` with a model list.

Use this request as a quick auth smoke test, not as a full key diagnosis. `429 insufficient_quota` belongs to billable routes such as `POST /v1/chat/completions`, and `503 service_unavailable` here means the model catalog is temporarily unavailable, not that the Bearer header is wrong.

## Bearer header format
**Request Example**
```
Authorization: Bearer gp-your-api-key
```

Keep the key in a server-side runtime or secret manager. Do not ship it in browser code, mobile bundles, or public repos.

## Store and rotate the key

- Save the full secret when it is first shown, because GonkaGate does not reveal it again later.
- Store the key outside source control, preferably in environment variables or a secret manager.
- Rotate by creating the replacement key first, moving traffic, then disabling the old key.
- If a key may be exposed, stop using it and rotate it immediately.

## Common first auth responses

| Response | What it means | Do this first |
|---|---|---|
| 401 invalid_api_key | The Bearer header is missing, malformed, copied incorrectly, or tied to a key or account state that cannot authenticate. | Check the Bearer header first, then confirm the key and account state. |
| 403 account_deactivated | API access is blocked by account state. | Restore the account state before retrying. |
| 429 insufficient_quota | The available prepaid USD balance is too low for a billable request such as `POST /v1/chat/completions`. It is not the expected result of `GET /v1/models`. | Treat it as billing and balance triage, not as a Bearer-header retry case. |
| 429 rate_limit_exceeded | You hit either a temporary rate limit or a configured limit on the API key itself. | First separate temporary throttling from a key limit. Retry only helps in the first case. |
| 503 service_unavailable | On `GET /v1/models`, model-catalog loading is temporarily unavailable. This is not a Bearer header or key-format problem. | Triage it as availability, then retry later or check support/status instead of rotating the key. |
Need exact diagnostics for the current key? Use [API key WhoAmI](/en/docs/api/api-reference/whoami/get-whoami).

## See also

- [Quickstart](/en/docs/quickstart) to send your first `chat.completions` request after authentication works.
- [Rate Limit Handling](/en/docs/api/reference/rate-limits) for current `/v1` request, token, and concurrency limits.
- [GonkaGate API Error Handling](/en/docs/api/reference/error-handling) for retry vs stop decisions after auth succeeds.
- [Pricing](/en/pricing) if the real problem is `insufficient_quota`, not invalid authentication.