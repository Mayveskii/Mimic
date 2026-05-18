# Network Domain — Artifacts

How network processes are stored as mesh slots.

---

## Slot Structure for Network Patterns

| Field | Value |
|---|---|
| domain | `DOMAIN_NETWORK` (3) |
| layer | `LAYER_PROCESS` (2) |
| modality | `MODALITY_CODE` (0) |
| pattern_name | `http_get_post` / `tcp_managed` / `websocket_upgrade` |
| invariants | `["url_validated", "rate_limited", "timeout_enforced", "size_bounded", "request_logged", "credential_pooled", "retry_transient_only", "dns_resolved", "fd_tracked", "websocket_validated"]` |
| polarity | `POLARITY_POSITIVE` (0) |
| survival_index | From successful API call patterns across repos |
| z_density | High (network patterns are common, well-compressed) |

---

## Pattern Codes

### http_get_post

```c
OpPacket chain[3] = {
    {OP_ORCH_VALIDATE,  .args = {{"url", ""}}},  // URL validation (SSRF check)
    {OP_SESS_BUDGET_CHECK, .args = {}},  // rate limit check
    {OP_NET_HTTP_GET,   .args = {{"url", ""}, {"timeout_ms", "30000"}}}
};
```

### tcp_managed

```c
OpPacket chain[4] = {
    {OP_ORCH_VALIDATE,    .args = {{"host", ""}, {"port", ""}}},
    {OP_NET_TCP_CONNECT,  .args = {{"host", ""}, {"port", ""}, {"timeout_ms", "10000"}}},
    {OP_NET_TCP_SEND,     .args = {{"fd", ""}, {"data", ""}}},
    {OP_NET_TCP_CLOSE,    .args = {{"fd", ""}}}
};
```

---

## Anti-Pattern Slots

| Anti-Pattern | Slot Name | counter_slot_id |
|---|---|---|
| AP-06 (undifferentiated retry) | `net_retry_all_errors` | `http_get_post` |
| AP-14 (infinite wait) | `net_no_timeout` | `tcp_managed` |
| AP-19 (toolloop eating iterations) | `net_rate_limit_retry` | `http_get_post` |

---

## Retrieval Path

Network patterns retrieved via:
1. Linear: exact name match for common patterns.
2. Keyword: invariant_hash for `timeout_enforced`, `url_validated`.
3. Semantic: "how do I safely call an API?" → domain=network.
