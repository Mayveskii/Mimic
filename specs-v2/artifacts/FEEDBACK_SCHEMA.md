# Feedback Schema — Execution Feedback Loop

Feedback records the outcome of applying a slot pattern in a real execution context. It is the mechanism by which the mesh learns and self-corrects.

Without feedback, the mesh is static. With feedback, survival indices update, patterns are promoted or demoted, and new slots are created from novel successful executions.

---

## JSON Feedback Structure

```json
{
  "feedback_id": "fb-uuid-v4",
  "version": 1,
  "created_at": "2026-05-17T12:05:00Z",
  
  "session_id": "sess-uuid-v4",
  "chain_id": 42,
  "slot_id": 18446744073709551615,
  
  "application": {
    "context_summary": "git atomic_commit on repo /workspace/mimic",
    "arguments": {
      "files": ["specs-v2/c-core/OPCODE_SPEC.md"],
      "message": "feat: add exact opcode enumeration"
    },
    "timestamp_start_ns": 1715923500000000000,
    "timestamp_end_ns": 1715923501450000000
  },
  
  "outcome": {
    "status": "success",
    "result_summary": "commit abc1234 created, 1 file changed, 4 insertions",
    "latency_ms": 145,
    "resource_usage": {
      "tokens_consumed": 8,
      "memory_peak_bytes": 4096,
      "energy_cost": 40.0
    }
  },
  
  "verification": {
    "method": "implicit",
    "verifier_count": 0,
    "votes": null
  },
  
  "post_conditions": {
    "invariant_checks": [
      {"invariant": "status_clean_after_commit", "result": "pass"}
    ],
    "state_hash_match": true
  },
  
  "anomalies": [],
  
  "novelty": {
    "is_novel": false,
    "similar_slots": [18446744073709551614, 18446744073709551613],
    "similarity_scores": [0.95, 0.88]
  }
}
```

---

## Required Fields

| Field | Type | Required | Description |
|---|---|---|---|
| feedback_id | UUID | yes | Unique feedback record |
| session_id | UUID | yes | Session that produced this feedback |
| chain_id | uint32 | yes | Chain ID within session |
| slot_id | uint64 | yes | Slot that was applied |
| application | object | yes | Context of application |
| outcome | object | yes | Result of application |

---

## Outcome Status Values

| Status | Meaning |
|---|---|
| success | Pattern applied, all invariants held |
| partial | Pattern applied, some invariants failed but no corruption |
| failure | Pattern failed, state restored via rollback |
| unexpected | Pattern succeeded but produced unexpected side effects |
| timeout | Pattern did not complete within timeout |
| skipped | Pattern was planned but not executed (validation rejected) |

---

## Verification Methods

| Method | Description |
|---|---|
| implicit | No explicit verification. Success assumed from result_code == 0 and invariant checks. |
| single | One verifier checked output. Used for non-critical ops. |
| double | Two independent verifiers (Verifier A correctness, Verifier B invariants). Both must pass. |
| manual | Human operator verified. Overrides automatic assessment. |

For CRITICAL (safety level 0) operations, method MUST be "double".
For DANGEROUS (safety level 1) operations, method MUST be "single" or "double".
For MUTATING (safety level 2) operations, method MAY be "implicit".
For SAFE (safety level 3) operations, method MAY be "implicit".

---

## Anomaly Detection

```json
{
  "anomalies": [
    {
      "type": "latency_spike",
      "severity": "warning",
      "expected_ms": 50,
      "actual_ms": 145,
      "description": "Commit latency 2.9x above baseline. Possible disk pressure."
    }
  ]
}
```

Anomaly types:
- `latency_spike`: Actual latency > 2x expected.
- `resource_leak`: FDs or mmaps left open after chain completion.
- `state_divergence`: Post-state hash does not match expected.
- `invariant_violation`: Post-condition check failed.
- `energy_overrun`: Actual energy > 2x estimated.
- `novel_pattern`: No similar slot found in mesh (similarity < 0.5).

---

## Novelty Assessment

After successful application, compare execution trace to existing slots:

```
If max(similarity_scores) < 0.5:
    → is_novel = true
    → Create FEEDBACK slot (SLOT_FLAG_FROM_FEEDBACK)
    → Queue for distillation review
Else if max(similarity_scores) > 0.9:
    → Increment usage_frequency of matching slot
    → Update survival_index with new data point
Else:
    → Record feedback under best-matching slot
    → Update z_density (compression may improve with more data)
```

---

## Feedback Processing Pipeline

```
FEEDBACK RECEIVED
    ↓
VALIDATE (check all required fields present, slot_id exists)
    ↓
ASSESS OUTCOME
    success → increment slot.success_count
    failure → increment slot.failure_count, check failure threshold
    timeout → mark slot as "flaky", reduce survival_index by 0.01
    ↓
DETECT ANOMALIES (compare to historical baselines)
    ↓
UPDATE INDICES
    success_count, failure_count, usage_frequency updated in slot header
    retrieval_count already incremented at retrieval time
    ↓
TRIGGER REVIEW (if needed)
    failure_rate > 0.2 → flag for manual review
    latency trend increasing → flag for optimization review
    anomaly count > 3 per session → flag for session audit
```

---

## Feedback to Slot Creation

When `is_novel == true` and `status == "success"`:

1. Create draft slot from execution trace.
2. Run distillation pipeline on trace (extract pattern, compute invariants).
3. QAC assessment on draft slot.
4. If precision ≥ 0.8: index as new slot with `SLOT_FLAG_FROM_FEEDBACK`.
5. If precision < 0.8: archive in feedback log, do not index.
6. Notify session that novel pattern was discovered and stored.

This is how the mesh grows: every novel successful execution contributes to the knowledge base.

---

## Feedback Retention

- Feedback records are stored for 90 days in hot storage (mmap-backed).
- After 90 days, aggregated into weekly rollup statistics (survival_index trends, failure_rate trends).
- Rollup statistics retained indefinitely in cold storage (compressed JSONL).
- Individual feedback records deleted after rollup to save space.

---

## Binary Feedback Record

For high-throughput storage, feedback is also stored in binary format:

```c
typedef struct {
    uint64_t feedback_id;        // Hash of UUID (first 8 bytes)
    uint64_t session_id;
    uint32_t chain_id;
    uint64_t slot_id;
    uint64_t timestamp_ns;
    uint8_t outcome_status;        // 0=success, 1=partial, 2=failure, 3=unexpected, 4=timeout, 5=skipped
    uint16_t latency_ms;
    float energy_cost;
    uint8_t anomaly_count;       // Up to 8 anomalies stored inline
    uint64_t anomaly_offset;     // Offset to anomaly details if > 8
    uint8_t padding[8];
} FeedbackRecord;
```

Size: 8+8+4+8+8+1+2+4+1+8+8 = 62 bytes + padding to 64.
1M feedback records = 64MB. Fits in L3 cache with mmap.
