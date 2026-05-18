# Network Domain — Invariants

Rules that MUST hold for every process in the network domain.

---

## NINV-01: URL Validated Before Request

**What it prevents:** SSRF attacks, internal network scanning, metadata theft.

**What it requires:** Before OP_NET_HTTP_GET/POST/CONNECT:
1. Scheme must be https (http allowed only for whitelisted domains).
2. Host must not resolve to private IP ranges (RFC 1918, loopback, link-local, CGN).
3. Host must not be 169.254.169.254 (cloud metadata endpoint).
4. DNS resolution must return public IP. No DNS rebinding (validate before request).

**Source of this rule:**
- graphify: SSRF protection with private-IP block, metadata block, DNS rebinding guard.
- Standard security practice.

**Consequence of violation:** Request REJECTED with `ERR_PERMISSION_DENY`. Reason: `private_ip`, `cloud_metadata`, `invalid_scheme`.

---

## NINV-02: Rate Limit Enforced

**What it prevents:** API abuse, account suspension, iteration budget exhaustion.

**What it requires:** Per-domain rate limit tracking. Default: max 60 requests/minute per domain. Configurable per API key. Exceeded → queue or reject with `retry_after`.

**Source of this rule:**
- hermes-agent: error classification includes rate-limit as retryable.
- AP-19 (toolloop eating iterations): 429 decrements iteration counter.

**Consequence of violation:** Request REJECTED with `rate_limited` and `retry_after: N` seconds.

---

## NINV-03: Timeout Always Set

**What it prevents:** Infinite hangs, dead sessions, wasted budget.

**What it requires:** Every network operation MUST have timeout_ms > 0. Default: 30000ms (30s). Max: 300000ms (5min). Zero timeout = REJECTED at validation.

**Source of this rule:**
- AP-14 (infinite wait): no timeout on streaming → hangs forever.
- AP-30 (missing cancellation boundary): long operation blocks forever.
- gastown: streaming health with 90s stale-stream detection.

**Consequence of violation:** Validation REJECTS with `ERR_INVALID_ARG`. If timeout fires during execution → `ERR_TIMEOUT`.

---

## NINV-04: Response Size Bounded

**What it prevents:** Memory exhaustion, OOM from large downloads.

**What it requires:** Max response size: 50MB binary, 10MB text. Exceeded → stream to temp file, truncate, or abort. Body truncated at limit with `truncated: true` flag.

**Source of this rule:**
- graphify: streaming fetch with size caps.
- Standard DoS protection.

**Consequence of violation:** Response truncated at limit. Model receives `truncated: true` with actual size. If binary > 50MB, abort with `ERR_OOM`.

---

## NINV-05: Every Request Logged

**What it prevents:** Unattributed requests, audit gaps, abuse.

**What it requires:** Every network request logged: timestamp, URL (host only, path redacted for sensitive APIs), method, status_code, duration_ms, size, session_id. Full URL logged only for internal/debugging endpoints.

**Source of this rule:**
- bun PR #30412: session enrichment on every tool use.
- Standard audit compliance.

**Consequence of violation:** Missing log entry detected at RESPOND phase. Warning: "network request without log entry".

---

## NINV-06: Credential Pool for Auth

**What it prevents:** Hardcoded API keys in requests, secret leakage in logs.

**What it requires:** Authenticated requests use credential pool (key_id reference), never inline keys. Credential pool rotates keys. Keys never appear in logs or session context.

**Source of this rule:**
- AP-07 (hardcoded secrets): API keys committed to source.
- hermes-agent: multi-key rotation, env fallback.
- caveman: sensitive path protection.

**Consequence of violation:** Request REJECTED if key detected in request body/URL. Log alert: "credential leaked in request".

---

## NINV-07: Retry Only on Transient Errors

**What it prevents:** Wasted retries on permanent failures, infinite loops.

**What it requires:** Retry only for: timeout, 5xx server errors, rate-limit (429 with retry-after), DNS resolution failure. NO retry for: 4xx client errors (400, 401, 403, 404), auth failure, SSL errors.

**Source of this rule:**
- hermes-agent: error classification (retryable vs permanent vs auth vs rate-limit).
- AP-06 (undifferentiated retry): 429 treated as 5xx → iteration budget consumed.

**Consequence of violation:** Wrong retry decision → `retry_count` decremented without progress. After 3 wrong retries → `ERR_PERMISSION_DENY` (circuit break on denial).

---

## NINV-08: DNS Resolution Before Connection

**What it prevents:** DNS rebinding attacks, TOCTOU between DNS and connect.

**What it requires:** Resolve hostname to IP BEFORE opening socket. Validate IP is public. Connect using resolved IP, not re-resolved hostname.

**Source of this rule:**
- graphify: DNS rebinding guard.
- Standard SSRF protection.

**Consequence of violation:** If resolved IP is private → REJECTED. If resolution fails → `ERR_NETWORK` (retryable).

---

## NINV-09: Connection Lifecycle Tracked

**What it prevents:** FD leaks, resource exhaustion.

**What it requires:** Every OP_NET_TCP_CONNECT creates a tracked FD in ExecContext. OP_NET_TCP_CLOSE MUST be called or FD auto-closed on chain completion. Max concurrent connections: 100 per session.

**Source of this rule:**
- EXEC_CONTEXT_SPEC.md: FD tracking.
- Resource management best practices.

**Consequence of violation:** FD leak detected at chain completion. Warning logged. If > 100 concurrent → `ERR_PERMISSION_DENY`.

---

## NINV-10: WebSocket Upgrade Validated

**What it prevents:** Hijacking, protocol confusion.

**What it requires:** OP_NET_WEBSocket validates upgrade response: HTTP 101 status, correct Sec-WebSocket-Accept hash, no unexpected headers.

**Source of this rule:**
- RFC 6455.
- Standard WebSocket security.

**Consequence of violation:** Upgrade REJECTED with `ERR_NETWORK` and "invalid websocket handshake".
