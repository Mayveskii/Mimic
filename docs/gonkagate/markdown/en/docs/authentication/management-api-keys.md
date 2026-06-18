# Management API Keys

Manage regular GonkaGate API keys and USD spend limits through /api/v1/keys, then understand the runtime rate limits that apply when those keys call /v1.

Use a `gpm-...` management API key with `/api/v1/keys` to list, create, update, and delete regular `gp-...` API keys that are available through this API. Management keys are for control-plane automation in the same GonkaGate account. They do not authenticate `/v1/*` model requests.

## Use the right key type

- Regular `gp-...` API keys authenticate `/v1/*` model requests.
- Management `gpm-...` API keys authenticate `/api/v1/keys*` routes.
- Management keys can only manage regular API keys owned by the same GonkaGate account that are available through `/api/v1/keys`.
- `read_write` keys can list, create, get, update, and delete those keys.
- `read_only` keys can only list and get individual keys.

Use management keys when you need to:

- create one compute API key per customer, workspace, or environment
- rotate or revoke regular API keys from automation
- apply per-key USD limits
- separate staging, production, CI, and internal traffic

## Create the management key once

1. Sign in to the dashboard.
2. Open [Dashboard → Settings → Management Keys](/en/dashboard/settings/management-keys).
3. Click **Create**.
4. Enter a display name and confirm.
5. Copy the `gpm-...` secret and store it securely.

Important rules:

- The current dashboard create flow only asks for a display name. It does not expose access or expiration controls.
- GonkaGate shows the full `gpm-...` secret only once.
- The account must have a verified email to create management keys and to use `/api/v1/keys`.
- Up to `25` active management keys are allowed per account.
- Create separate management keys for production, staging, CI, and internal tooling instead of sharing one credential.

## Call `/api/v1/keys` with a `gpm-...` key

Use this base URL in GonkaGate production:
**Endpoint**
```
https://api.gonkagate.com/api/v1/keys
```

On another deployment, keep `/api/v1/keys` and replace only the host.

Send the management key in the Bearer header:
 Send the management key in the Bearer header
```
Authorization: Bearer gpm-YOUR_MANAGEMENT_KEY
```

## Supported operations

| Method | Route | What it does |
|---|---|---|
| GET | /api/v1/keys | List up to 100 regular API keys available through this API. |
| POST | /api/v1/keys | Create a regular API key and return the raw gp-... secret once. |
| GET | /api/v1/keys/:hash | Get one key by its external hash. |
| PATCH | /api/v1/keys/:hash | Rename a key, disable or re-enable it, change limit settings, or change expiration. |
| DELETE | /api/v1/keys/:hash | Delete a key by its external hash. |
`hash` is the `64`-character identifier returned in `data.hash`. It is not the raw `gp-...` key value.

## Manage regular keys programmatically
**TypeScriptPythoncurlTypeScript Example**
```
const MANAGEMENT_API_KEY = "gpm-YOUR_MANAGEMENT_KEY";
const BASE_URL = "https://api.gonkagate.com/api/v1/keys";

const headers = {
Authorization: `Bearer ${MANAGEMENT_API_KEY}`,
"Content-Type": "application/json",
};

const listResponse = await fetch(BASE_URL, { headers });

const listWithDisabledResponse = await fetch(
`${BASE_URL}?offset=100&include_disabled=true`,
{ headers },
);

const createResponse = await fetch(BASE_URL, {
method: "POST",
headers,
body: JSON.stringify({
name: "Customer Production Key",
limit: 50,
limit_reset: "monthly",
expires_at: "2026-12-31T23:59:59Z",
}),
});

const created = await createResponse.json();
const keyHash = created.data.hash;

const getResponse = await fetch(`${BASE_URL}/${keyHash}`, { headers });

const updateResponse = await fetch(`${BASE_URL}/${keyHash}`, {
method: "PATCH",
headers,
body: JSON.stringify({
name: "Customer Production Key v2",
disabled: true,
limit: null,
expires_at: null,
}),
});

const deleteResponse = await fetch(`${BASE_URL}/${keyHash}`, {
method: "DELETE",
headers,
});
```

On create, save both values:

- `key` is the one-time raw `gp-...` secret for the new regular API key.
- `data.hash` is the `64`-character external identifier used by `GET`, `PATCH`, and `DELETE`.

## Request rules that change behavior

| Topic | Rule |
|---|---|
| Listing | offset defaults to 0, max is 10000, and include_disabled defaults to false. Responses are ordered from newest to oldest. |
| Limits | limit is a positive USD amount. limit_reset supports daily, weekly, or monthly. If you set limit and omit limit_reset, the key gets a lifetime total limit. |
| Limit validation | limit_reset without limit returns 400. Send limit: null on PATCH to remove the limit and clear limit_reset. |
| Expiration | expires_at must be a future RFC 3339 / ISO 8601 timestamp with timezone. Send expires_at: null on PATCH to remove expiration. |
| Access mode | read_only management keys can call GET routes only. Write operations return 403 Management API key is read-only. |
| Response format | Successful /api/v1/keys responses use OpenRouter-style field names. Errors still use the standard GonkaGate /api/* error envelope. |
| BYOK | include_byok_in_limit is intentionally unsupported and returns 400. |

## Runtime limits for keys created here

`limit` is a USD spend cap, not a request-throughput limit. A regular `gp-...` key created through `/api/v1/keys` is checked against both its own key bucket and the owning account bucket when it is used on `/v1/*` model requests. If any bucket is exhausted, the request can return `429 rate_limit_exceeded`.

Current GonkaGate production limits for authenticated `/v1/*` model requests:

| Scope | Request and token limits | Burst and concurrency |
|---|---|---|
| Regular API key (gp-...) | 120 RPM and 5,000,000 TPM | 40 request burst and 10 concurrent requests |
| Regular API key + source IP | 120 RPM and 5,000,000 TPM | 40 request burst |
| Owning account | 12,000 RPM and 300,000,000 TPM | 4,000 request burst and 1,000 concurrent requests |
| Source IP | 12,000 RPM and 300,000,000 TPM | 4,000 request burst |
| Distinct regular keys from one source IP | 5,000 keys per hour | Applies when many customer keys exit through the same backend IP |
RPM means requests per minute. TPM means estimated tokens per minute.

For a 500-1000 customer-key setup under one GonkaGate account, each customer key gets its own per-key bucket, but all of those keys still share the owning account bucket and the account prepaid USD balance. Management keys do not create separate `/v1/*` runtime buckets.

## What successful calls return

Successful `/api/v1/keys` calls use OpenRouter-style field names:

- `GET /api/v1/keys`: `{ data: OpenRouterKey[] }`
- `GET` and `PATCH /api/v1/keys/:hash`: `{ data: OpenRouterKey }`
- `POST /api/v1/keys`: `{ key: "gp-...", data: OpenRouterKey }`
- `DELETE /api/v1/keys/:hash`: `{ deleted: true }`

All `limit*` and `usage*` values are USD.
 What successful calls return
```
{
  "hash": "4f2f5b6f0b0e16eab8c2e55d4f2f0ce4c6f5b91f2a28f8e93b0f8c3b8b112233",
  "label": "gp-AbCd...WXYZ",
  "name": "Customer Production Key",
  "disabled": false,
  "limit": 50,
  "limit_remaining": 50,
  "limit_reset": "monthly",
  "usage": 0,
  "usage_daily": 0,
  "usage_weekly": 0,
  "usage_monthly": 0,
  "created_at": "2026-03-16T10:00:00.000Z",
  "updated_at": "2026-03-16T10:00:00.000Z",
  "expires_at": "2026-12-31T23:59:59.000Z"
}
```

## Common failures

Failed requests keep the standard GonkaGate `/api/*` error envelope, even on the OpenRouter-compatible route:
**Error Example**
```
{
  "success": false,
  "error": {
    "message": "Management API key is read-only",
    "statusCode": 403,
    "requestId": "req_abc123",
    "timestamp": "2026-03-16T10:00:00.000Z"
  }
}
```

| Response | What to do first |
|---|---|
| 401 Unauthorized | Check that the Bearer header uses a valid active gpm-... key, not a regular gp-... key. This route returns 401 for missing, unknown, disabled, expired, and other invalid management keys. |
| 403 Management API key is read-only | Retry with a read_write management key for POST, PATCH, or DELETE. |
| 403 Please verify your email address to access this feature | Verify the email on the owning account before retrying. |
| 400 limit_reset requires limit | Send limit whenever you send limit_reset. |
| 400 expires_at must be ... or BYOK fields are not supported | Fix the payload and resend it. Retrying the same payload will fail again. |
| 404 API key not found | Check the data.hash value. The key may have been deleted or may belong to another account. |

## See also

- [API Error Handling](/en/docs/api/reference/error-handling) when this automation moves into production and you need retry boundaries, request IDs, and support-ready failure handling.
- [Authentication and API Keys](/en/docs/authentication/api-keys) for regular `gp-...` keys used on `/v1/*`.
- [Quickstart](/en/docs/quickstart) if you need the first model request rather than control-plane automation.
- [API Reference Overview](/en/docs/api/reference/overview) for the public `/v1` request and response contract after keys are provisioned.