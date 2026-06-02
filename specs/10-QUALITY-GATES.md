# QUALITY-GATES.md — Mimic

Quality Axis Conditions (QAC) that every artifact, operation, and decision must satisfy before entering storage, execution, or mesh. Measured thresholds derived from binary-mesh production data and source history.

---

## 1. QAC Definitions

### QAC-1: Survival Index Provenance

```
condition: survival_index MUST be computed from git blame, not guessed
measurement: git blame surviving_lines / total_lines_added on source_commit
threshold: survival_index >= 0.7 for slot candidate, < 0.1 = discard
```

Violation consequence: Fake SI → model trusts unproven patterns → executes bad chains → L0 Stability degrades.

Measured baseline (binary-mesh L0): 8000/10000 (5585 samples). Threshold for alert: L0 < 7000.

---

### QAC-2: Invariant Coverage

```
condition: Every artifact MUST have >= 1 verifiable invariant
measurement: invariant_count >= 1 AND invariant.verification is not empty
threshold: artifact_precision = survival_index * (verified_invariants / required_invariants) * extraction_reproducibility
```

Violation consequence: No invariant → mimicry control fails (BEHAVIOR.md #10) → model guesses → L6 Reuse degrades.

Measured baseline (binary-mesh L6): 1931/10000 (428 samples). Threshold for alert: L6 < 1500.

SuperInvariant evidence (from enricher, 148121 slots across 8 domains):
- "nullable reference must be guarded before dereference" — 9409 members (security)
- "incorrect logic must be replaced to satisfy postcondition" — 57724 members (logic)
- "missing precondition check must be added before operation" — 2491 members (trace)

---

### QAC-3: Energy Cost Measurement

```
condition: Energy cost MUST be measured, not estimated
measurement: ops_calculate_action() with actual cost_tokens * cost_time_us from exec latency_ns
threshold: estimated_energy / measured_energy <= 1.2 (20% tolerance)
```

Violation consequence: Wrong energy cost → budget overflow → chain aborts mid-exec → L9 Completion degrades.

Measured baseline (binary-mesh L9): 7072/10000 (50 samples). Trend: declining from 10000. Threshold for alert: L9 < 6000.

Current ops.c measured costs (from ops_register_builtins):

| OpCode | cost_tokens | cost_time_us | cost_memory_bytes |
|--------|------------|-------------|-------------------|
| OP_NOP | 0.0 | 0.01 | 0.0 |
| OP_SYS_FILE_EXISTS | 1.0 | 10.0 | 0.0 |
| OP_SYS_DIR_CREATE | 2.0 | 50.0 | 4096.0 |

Action required: remaining 43 OpCodes need measured costs. Current Go matrix (energy_cost_matrix.go) uses estimated defaults — must be replaced with measured values from binary-mesh execution traces.

---

### QAC-4: Conflict Matrix Derivation

```
condition: Conflict matrix entries MUST be derived from observed conflicts, not invented
measurement: Each conflict rule MUST reference a source commit or SuperInvariant
threshold: conflict_matrix coverage = documented_rules / total_possible_conflicts >= 0.05
```

Violation consequence: False conflict → blocks valid chains. Missing conflict → allows invalid chains → L0 Stability degrades.

Current state (ops.c g_conflict_matrix): 1 rule (OP_SYS_EXEC x OP_SYS_EXEC). Go conflict_matrix.go: 4 rules. Required: minimum 46*45/2 = 1035 potential pairs, need >= 52 documented rules.

SuperInvariant conflict evidence (from enricher):
- "concurrent shared state requires lock before access" — 1936 members → WRITE x WRITE conflict
- "concurrent access requires synchronization" — 6 members → READ x WRITE without SYNC conflict

Historical conflict evidence (from binary-mesh git history):
- buildMessages mutation x buildMessages read = conflict (11 revisions, corruption c9e83c3)
- toolloop 429 iteration x budget check = conflict (675601e)
- goroutine-per-epoch prune x Store/Retrieve under mutex = conflict (d9c46cae0)

---

### QAC-5: Z-density Computation

```
condition: Z-density MUST be computed from actual slot data, not defaulted to 0
measurement: z_density_compute(slot) = (Σ survival_i * weight_i) / slot_volume
threshold: Z(slot) > 0 for any slot entering mesh
```

Violation consequence: Z-density = 0 → RAG signal 5 is useless → ranking degraded → L4 Usefulness degrades.

Measured baseline (binary-mesh L4): 2100/10000 (23 samples). Lowest quality axis. Threshold for alert: L4 < 1500.

---

### QAC-6: Decision Consistency

```
condition: Decision patterns MUST NOT contradict existing matrix entries
measurement: For each new decision artifact, check conflict_matrix for contradictions
threshold: zero contradictions allowed
```

Violation consequence: Contradictory decisions in same domain → model gets conflicting advice → L0 Stability + L4 Usefulness degrade.

---

### QAC-7: Artifact Precision

```
condition: artifact_precision > 0 (all three components non-zero) to enter deep cache
measurement: artifact_precision = survival_index * invariant_coverage * extraction_reproducibility
threshold: artifact_precision >= 0.8 for deep cache, < 0.8 = local only
```

Violation consequence: Low-precision artifact in deep cache → other models pull bad data → L6 Reuse degrades.

---

### QAC-8: Multimodal Integrity

```
condition: Multimodal extraction MUST verify modality-specific integrity
measurement: image hash match, diagram parse success, table schema conformance
threshold: 100% integrity for CODE/TEXT, >= 95% for IMAGE/DIAGRAM
```

Violation consequence: Corrupted image/diagram → model acts on wrong architecture understanding → L9 Completion degrades.

---

### QAC-9: Anti-Pattern Polarity (NEW)

```
condition: Every NEGATIVE artifact MUST link to a POSITIVE counter_pattern
measurement: For polarity=NEGATIVE, counter_pattern_id MUST be non-empty AND reference existing POSITIVE artifact
threshold: zero orphaned NEGATIVE artifacts
```

Violation consequence: Orphaned negative = model knows what NOT to do but not what TO do → partial knowledge → L4 Usefulness degrades.

Rationale: From source history analysis — every successful pattern has a failure predecessor:
- ZFC (gastown) ← stale cache failures
- rollback-on-failure (gastown) ← break-on-failure (gonka 86c686d92)
- two-vote (bun) ← single-verifier insufficiency
- deny-first pipeline (code-mode) ← unvalidated input (gonka #304)
- error_classifier (hermes-agent) ← undifferentiated 429/5xx handling (binary-mesh 0e36abd)
- sync.Once (gonka 8d1cd00f3) ← panic recovery failure
- sequential prune (gonka d9c46cae0) ← goroutine-per-epoch under mutex

---

### QAC-10: Temporal Consistency (NEW)

```
condition: survival_index MUST be re-validated when blame data changes
measurement: On source_commit new_child_count > 0, re-compute survival_index within 24h
threshold: survival_index drift > 0.1 from recorded value → flag for re-validation
```

Violation consequence: Stale SI → artifact appears more durable than it is → model trusts code that was later deleted → L0 Stability degrades.

---

### QAC-11: Cross-Domain Conflict (NEW)

```
condition: Conflict matrix MUST include cross-domain pairs
measurement: At least one conflict rule per domain-pair where resource_bitmask overlaps
threshold: all (8 domains * 7 pairs) / 2 = 28 cross-domain pairs have at least 1 rule
```

Violation consequence: Cross-domain conflict undetected → e.g. git operation conflicts with network operation → L8 Latency degrades.

Measured baseline (binary-mesh L8): 8005/10000 (4319 samples). Stable. Threshold for alert: L8 < 7500.

---

### QAC-12: Provenance Chain (NEW)

```
condition: Every artifact MUST carry extraction_hash = sha256(extractor_version + parameters)
measurement: extraction_hash present AND matches current extractor version
threshold: extraction_hash drift (different version) → re-extract, do not use cached
```

Violation consequence: Artifact extracted with old tool → different output with new tool → non-reproducible → QAC-1 through QAC-8 all potentially violated.

---

### QAC-13: Revert Detection (NEW)

```
condition: Commits with "Revert ..." message MUST generate NEGATIVE artifact automatically
measurement: git log --grep="Revert" → for each, create NEGATIVE artifact with decision_survival=0.0
threshold: 100% of explicit revert commits within source repos become NEGATIVE artifacts
```

Violation consequence: Revert = active negative knowledge discarded. Current distillation treats survival < 0.1 as discard, but revert is explicitly different — someone actively removed code, not just code that faded.

Historical revert evidence:
- binary-mesh 704beab: Revert "fix: set max_tokens=2048" (402 from proxy)
- binary-mesh c03b244/87c27d8: flip-flop "forbid shell exploration" (2 min apart)
- binary-mesh b6d7907: Revert "workspace snapshot in system context"
- gonka 80b2598b8: Revert "regenerate seed (#375)"
- next.js #92320: Revert "simplify session dependent tasks and add TTL support" (from enricher)
- cilium #45098: Revert "bgp: Ensure ServerLogger uses BGP instance name" (from enricher)

---

## 2. Meta-Invariant

```
no_side_effect_without_prior_validation
```

Every side effect (I/O, state mutation, resource allocation, network call) MUST be preceded by validation that the operation is safe, within budget, and not conflicting. This single invariant unifies all 13 QACs and is the root cause of every historical failure:

| Failure | Missing Validation | QAC Violated |
|---------|-------------------|--------------|
| break-on-failure → unbounded disk (gonka) | No validation that PruneEpoch succeeded before advancing | QAC-1, QAC-2 |
| goroutine-per-epoch under mutex (gonka) | No validation that mutex was free before spawning | QAC-4, QAC-11 |
| context injection corruption (binary-mesh) | No validation that buildMessages structure was stable | QAC-6, QAC-10 |
| panic on negative coins (gonka) | No validation of input before NewInt64Coin | QAC-2, QAC-7 |
| hardcoded API keys (binary-mesh) | No validation that secrets were not in source | QAC-8, QAC-12 |
| 429 not distinguished from 5xx (binary-mesh) | No classification before retry decision | QAC-3, QAC-6 |
| UnboundedQueue Close via panic recovery (gonka) | No idempotency validation | QAC-2, QAC-7 |
| index out of range in backoff (binary-mesh) | No bounds check before array access | QAC-2, QAC-5 |

---

## 3. Quality Gate Pipeline

Before ANY artifact enters storage:

```
1. VALIDATE_SOURCE
   - git blame exists for source_commit? (QAC-1)
   - Is this a Revert commit? → auto-generate NEGATIVE artifact (QAC-13)
   - extraction_hash matches current extractor? (QAC-12)

2. VALIDATE_COMPUTATION
   - survival_index formula correct? (QAC-1)
   - invariant parseable and verifiable? (QAC-2)
   - energy cost measured, not estimated? (QAC-3)
   - Z-density computed from actual data? (QAC-5)

3. VALIDATE_CONSISTENCY
   - conflicts with existing matrix? (QAC-4, QAC-11)
   - contradicts existing artifact in same domain? (QAC-6)
   - NEGATIVE artifact has counter_pattern? (QAC-9)

4. VALIDATE_PRECISION
   - all three precision components > 0? (QAC-7)
   - multimodal integrity verified? (QAC-8)
   - temporal consistency: SI drift check? (QAC-10)

5. STORE
   - All gates pass → compress → write → index
   - Any gate fail → mark as pending with failure reason → do NOT store
```

---

## 4. Quality Matrix Axis Mapping

| QAC | Primary Axis | Threshold (from binary-mesh data) | Alert Condition |
|-----|-------------|-----------------------------------|-----------------|
| QAC-1 | L0 Stability | SI >= 0.7 for slot, >= 0.1 for partial | L0 < 7000 |
| QAC-2 | L6 Reuse | >= 1 invariant per artifact | L6 < 1500 |
| QAC-3 | L9 Completion | estimated/measured <= 1.2 | L9 < 6000 |
| QAC-4 | L0 Stability | >= 52 documented conflict rules | conflict coverage < 5% |
| QAC-5 | L4 Usefulness | Z-density > 0 for all mesh slots | L4 < 1500 |
| QAC-6 | L0 + L4 | Zero contradictions | any contradiction found |
| QAC-7 | L6 Reuse | precision >= 0.8 for deep cache | L6 < 1500 |
| QAC-8 | L9 Completion | 100% CODE/TEXT, 95% IMAGE/DIAGRAM | L9 < 6000 |
| QAC-9 | L4 Usefulness | Zero orphaned NEGATIVE artifacts | any orphaned NEGATIVE |
| QAC-10 | L0 Stability | SI drift < 0.1 within 24h | drift >= 0.1 |
| QAC-11 | L8 Latency | 28 cross-domain pairs covered | L8 < 7500 |
| QAC-12 | All | extraction_hash present and current | any hash mismatch |
| QAC-13 | L4 + L6 | 100% revert commits → NEGATIVE | any missed revert |

Current composite score: 4784/10000. Research-level threshold: >= 8000.

---

## 5. Measured SuperInvariant Population (from binary-mesh enricher)

These SuperInvariants from 148121 slots across 8 domains directly feed QAC-2 and QAC-4:

| SuperInvariant | Members | Level | QAC Relevance |
|---------------|---------|-------|---------------|
| incorrect logic must be replaced to satisfy postcondition | 57724 | logic | QAC-2: universal invariant |
| nullable reference must be guarded before dereference | 9409 | security | QAC-2, QAC-4: nil-check conflict rule |
| new functionality requires explicit dependency import | 12146 | architecture | QAC-2: dependency invariant |
| behavioral change requires test coverage | 5910 | trace | QAC-2: test invariant |
| missing precondition check must be added before operation | 2491 | trace | QAC-2: precondition invariant |
| user input requires validation before processing | 2231 | trace | QAC-2, QAC-9: input validation anti-pattern |
| concurrent shared state requires lock before access | 1936 | security | QAC-4: concurrency conflict rule |
| error condition must be surfaced not swallowed | 1964 | logic | QAC-2: error handling invariant |
| fallible external call requires retry with backoff | 1943 | trace | QAC-3: retry energy cost |
| arithmetic operation requires overflow guard | 1140 | security | QAC-2: overflow invariant |
| collection access requires length verification | 1090 | trace | QAC-2: bounds check invariant |
| operation boundary requires observability signal | 819 | security | QAC-2: observability invariant |
| long-running operation requires cancellation boundary | 590 | security | QAC-3: timeout energy cost |
| concurrent access requires synchronization | 6 | security | QAC-4: race condition conflict |

---

## 6. Efficiency Metrics (from binary-mesh workspace)

### 6.1 Enricher Efficiency

```
api_calls_made = 72603
cycles_completed = 1765
prs_scanned = 33755
slots_created = 32949
rate_limit_hits = 0

efficiency = slots_created / api_calls_made = 32949 / 72603 = 0.454 (45.4%)
throughput = slots_created / cycles = 32949 / 1765 = 18.67 slots/cycle
pr_yield = slots_created / prs_scanned = 32949 / 33755 = 0.976 (97.6%)
```

### 6.2 RAG Efficiency

```
ast_symbols = 38829
kw_terms = 117230
md_entries = 6367
comments = 4979
qdrant_points = 180684

signal_ratio = (ast + kw + md + comment) / qdrant = (38829 + 117230 + 6367 + 4979) / 180684 = 0.924
```

### 6.3 Dual Solve Efficiency

```
total_comparisons = 8
avg_latency_ms ≈ 1,886,428 ms (31 min)
best_latency = 53528 ms
worst_latency = 5691125 ms
score_range = 0..46.5
patch_rate = 0/8 = 0% (all NO_PATCH or PENDING)
```

Efficiency problem: dual_solve produces no patches. Token cost is high (inference calls), latency is high (avg 31 min), yield is zero. Root cause: tasks too vague or approaches not sufficiently differentiated.

### 6.4 Context Efficiency

```
est_tokens = 19635 / 32768 max = 59.9% pressure
messages = 34
tokens_per_message = 19635 / 34 = 577.5
```

### 6.5 Quality Axis Efficiency

```
L0_samples = 5585 → 8000 score → 1.43 score per sample
L4_samples = 23 → 2100 score → 91.3 score per sample (HIGH — needs more samples)
L6_samples = 428 → 1931 score → 4.51 score per sample
L8_samples = 4319 → 8005 score → 1.85 score per sample
L9_samples = 50 → 7072 score → 141.4 score per sample (HIGH — needs more samples)

Under-sampled axes: L4 (Usefulness) and L9 (Completion) need more data points.
```

---

## 7. Distillation Efficiency Targets

Based on measured enricher throughput and quality axis baselines:

| Metric | Current | Target (Phase 2) | Target (Phase 4) |
|--------|---------|-------------------|-------------------|
| Slots created | 32949 | 50000 | 150000+ |
| L0 Stability | 8000 | 8500 | 9000+ |
| L4 Usefulness | 2100 | 4000 | 7000+ |
| L6 Reuse | 1931 | 3500 | 7000+ |
| L8 Latency | 8005 | 8500 | 9000+ |
| L9 Completion | 7072 | 8000 | 9000+ |
| Composite | 4784 | 6500 | 8500+ |
| Conflict rules | 5 | 52 | 200+ |
| Dual_solve patch_rate | 0% | 20% | 50%+ |
| Enricher efficiency | 45.4% | 60% | 75%+ |
| Anti-pattern coverage | 0 | 40+ | 200+ |

---

## 8. Ops Engine Integration

### 8.1 Conflict Matrix Population from QAC-4

Every conflict rule in the ops engine MUST reference its source:

```c
typedef struct {
    OpCode op1;
    OpCode op2;
    ConflictLevel level;
    const char* source_commit;
    const char* superinvariant;
} ConflictRuleEvidence;

// Example: from binary-mesh history
ConflictRuleEvidence rule_buildMessages = {
    .op1 = OP_IO_WRITE,
    .op2 = OP_IO_READ,
    .level = ConflictHigh,
    .source_commit = "c9e83c3",
    .superinvariant = "concurrent shared state requires lock before access"
};
```

### 8.2 Energy Cost Correction from QAC-3

Current BuildDefaultEnergyMatrix() uses hardcoded estimates. Must be replaced:

```go
func BuildMeasuredEnergyMatrix(metrics []OpMetric) *EnergyCostMatrix {
    em := NewEnergyCostMatrix()
    for _, m := range metrics {
        em.Set(m.Opcode, m.MeasuredTokenCost, m.MeasuredLatencyMs)
    }
    return em
}
```

### 8.3 Validation Before Execution (Meta-Invariant)

Already implemented in ops_execute_chain (validate before execute). The gap is in the QUALITY of validation rules, not the pattern itself. QAC-4 and QAC-11 define what must be added.
