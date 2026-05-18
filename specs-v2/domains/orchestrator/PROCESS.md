# Orchestrator Domain — Model Amplification Engine

How any model, however weak, becomes strong by delegating execution to Mimic.

---

## What This Domain Does

The orchestrator is the amplification layer. A model expresses intent — the orchestrator transforms that intent into a validated, tokenized, measured process that the model follows step by step. The model does not need to know git, build systems, file operations, or network protocols. The orchestrator provides the proven process; the model provides the goal.

The orchestrator is not a generic workflow engine. It is a behavior synthesizer: it takes proven patterns from mesh slots, validates them against current context, assembles them into a coherent chain, enforces hard constraints, measures cost, and returns the result with full traceability.

---

## Processes

### intent_to_validated_chain

**When to use:**  
Every time a model calls Mimic with any task. This is the entry point for all model-Mimic interaction.

**Goal:**  
Transform model intent into a validated OpPacket chain ready for deterministic execution.

**Chain (semantically):**

1. **CLASSIFY** — understand what the model wants.
   - Determine task domain (git, build, io, network, system, analysis, mixed).
   - Determine safety level (0=critical/destructive, 1=dangerous, 2=safe with side effects, 3=read-only).
   - Determine task type (read-only query, safe mutation, git workflow, build pipeline, deploy, pattern application, parallel ops, RAG search, multi-pipeline, workspace self-build).
   - Identify named scenario match or mark as custom chain.
   - # UNCERTAIN: classification confidence threshold needs calibration on real model intents

2. **PLAN** — build the tokenized process.
   - If named scenario matches: instantiate scenario template with model-provided arguments.
   - If custom: build chain from model-provided explicit OpPacket sequence.
   - Query mesh slots for relevant patterns in classified domain.
   - Incorporate matching patterns if preconditions pass (mimicry control: check preconditions against current context; if pass → include; if fail → degrade to next_lower_tier).
   - Estimate total energy cost.
   - Apply borrowed behaviors (phase graph structure, permission pipeline, edit scope isolation).
   - Compute resource_bitmask for cross-pipeline conflict detection.

3. **VALIDATE** — verify the chain before ANY execution.
   - Pairwise conflict check: for every pair (op_i, op_j) where i < j, check conflict_matrix[op_i][op_j].
     - If conflict = 1 → REJECT chain, return which pair conflicts and why.
   - Energy budget check: sum of energy costs ≤ session budget_remaining.
   - Permission check:
     - Safety level 0 → requires explicit allow + 2-vote verify.
     - Safety level 1 → requires auto-classifier pass or explicit allow.
     - OP_FLAG_DANGEROUS → always requires explicit allow.
     - 3 consecutive denials → circuit break, manual mode only.
   - Invariant check: verify all scenario invariants are satisfiable with current state.
   - # UNCERTAIN: how many conflict rules are sufficient for production use

4. **EXEC** — execute the validated chain.
   - Call ops_execute_chain via CGO bridge.
   - Each OpPacket executed sequentially within a single chain.
   - Latency measured per operation with CLOCK_MONOTONIC.
   - On failure:
     - If retry_count > 0 → retry with exponential backoff.
     - If operation is ATOMIC → rollback entire chain to pre-execution state.
     - If scenario defines explicit rollback → execute rollback chain.
     - Record failure type in session for error classification.
   - Track open resources (fd, mmap) in ExecContext.
   - On chain completion → close all tracked resources.

5. **VERIFY** — post-execution verification when required by scenario.
   - If scenario requires 2-vote verify:
     - Verifier A checks output correctness.
     - Verifier B checks invariant preservation.
     - Both pass → VERIFIED.
     - Both fail → REJECTED, rollback.
     - Disagreement → tiebreak via additional check or manual escalation.
   - If scenario does not require verify → SKIP.

6. **RESPOND** — return result to model.
   - Assemble response: status, artifacts, metrics, validation report, pattern references.
   - If WRITE operations occurred → trigger workspace index re-index.
   - If pattern is novel (not in mesh) → create feedback artifact for local mesh storage.
   - If artifact_precision ≥ 0.8 and survival confirmed → flag for deep cache.
   - Update session state: budget consumed, context accumulated, denial count maintained.
   - Compress session snapshot, verify hash.

**Hard constraints:**
- No EXEC without passed VALIDATE. This is absolute.
- No phase may be skipped. The chain goes through all six phases in order.
- Rollback: VALIDATE fail → PLAN gets failure context. EXEC fail → PLAN gets failure context. VERIFY fail → EXEC gets failure context.
- Budget exceeded at any point → immediate halt, model notified with exact consumption breakdown.
- Circuit break after 3 consecutive denials → manual mode only.

**Invariants:**
- CLASSIFY does not execute any operations. It only reads state.
- PLAN does not execute. It produces a plan that must be validated.
- VALIDATE does not execute. If validation fails → chain NEVER executed.
- EXEC only runs validated chains.
- VERIFY only runs on critical operations when specified.
- RESPOND always provides complete, honest result. No hidden failures.

**Result when successful:**
```
status: "success"
artifacts: <result of execution>
metrics: {
  total_latency_ns: <sum of op latencies>
  total_tokens_consumed: <estimated>
  memory_peak_bytes: <measured>
  energy_cost: <measured>
}
validation_report: {
  conflict_check: "passed"
  budget_check: "passed"
  permission_check: "passed"
  invariants_verified: ["..."]
}
pattern_references: [
  {domain: "git", pattern: "atomic_commit", survival_index: 0.85, z_density: 0.72}
]
```

**Result when failed:**
```
status: "failure" | "rejected" | "budget_exceeded"
reason: "conflict_pair: OP_X × OP_Y" | "budget: 15.2 > 10.0 remaining" | "permission_denied"
state_restored: true  # rollback executed if applicable
partial_result: <what was completed before failure>
```

**How a model uses this:**
Model says "commit these files safely" → orchestrator CLASSIFY="git domain, atomic_commit scenario" → PLAN instantiates atomic_commit template → VALIDATE checks conflicts, budget, permissions → EXEC runs status→diff→add→commit → VERIFY (not required for standard commit) → RESPOND returns commit hash + metrics.

Model does not see OpCodes. Model sees: "Here is the process. Step 1: check status. Step 2: review diff. Step 3: stage files. Step 4: commit." Model follows.

---

### mesh_query

**When to use:**  
Model needs proven patterns for a specific domain or task. "How do I handle X?" or "Find patterns for Y."

**Goal:**  
Retrieve ranked proven patterns from mesh with full provenance.

**Chain (semantically):**

1. Classify query domain and type.
2. Linear lookup: si_query_domain_layer(domain, layer) → exact match by domain:layer:modality:pattern_name.
3. If no exact match → keyword lookup: si_query_state_hash(invariant_hash) → all artifacts sharing same invariants.
4. If still insufficient → semantic search: int8_quantize(query) → batch_cosine_int8 → top-k → re-rank by survival_index × z_density.
5. Return ranked list with provenance (source repo, commit, survival index, z-density, polarity).

**Hard constraints:**
- RAG without survival signal = unverified. Every result must carry survival index.
- If polarity = NEGATIVE → must return linked COUNTER pattern (what TO do instead).
- If artifact_precision < 0.8 → result tagged "low_precision, use with caution".

**Invariants:**
- Every result has survival_index > 0.0.
- Every result has at least one verifiable invariant.
- NEGATIVE results never returned without COUNTER.

**Result when successful:**
```
status: "success"
results: [
  {
    pattern_name: "rollback_on_failure",
    domain: "resource_cleanup",
    survival_index: 0.92,
    z_density: 0.81,
    polarity: "POSITIVE",
    source: "gastown/internal/polecat/manager.go",
    invariants: ["partial_failure_leaves_no_orphans"]
  }
]
retrieval_path: "linear" | "keyword" | "semantic"
```

**Result when failed:**
```
status: "no_results"
domain: "distributed_consensus"
suggestion: "This domain is not yet distilled. Run distillation for: etcd, k8s, cockroachdb."
```

**How a model uses this:**
Model says "how do I safely clean up resources on failure?" → orchestrator queries mesh → returns rollback_on_failure pattern with 0.92 survival, from gastown, with invariant "partial_failure_leaves_no_orphans". Model applies this pattern exactly as specified, does not improvise.

---

### parallel_pipeline_orchestration

**When to use:**  
Model has multiple independent tasks that can run simultaneously.

**Goal:**  
Execute multiple pipelines concurrently with isolation, conflict detection, and budget sharing.

**Chain (semantically):**

1. For each submitted intent → run CLASSIFY and PLAN independently.
2. Cross-pipeline conflict check:
   - For every pair of pipelines → check resource_bitmask overlap.
   - Overlap → serialize (pipeline A first, then pipeline B).
   - No overlap → parallel (up to 10 concurrent pipelines).
3. Shared resources (git index, build cache) → explicit locks.
4. Each pipeline follows full 6-phase sequence independently.
5. If any pipeline fails → that pipeline rolls back independently.
   - Other pipelines continue unless they depend on failed pipeline's output.

**Hard constraints:**
- No two pipelines write to the same resource simultaneously.
- Max 10 concurrent pipelines.
- Resource_bitmask overlap → always serialize, never allow race.

**Invariants:**
- Read-sharing is always allowed.
- Pipeline A failure does not affect pipeline B unless explicit dependency.
- Total budget across all pipelines ≤ session budget.

**Result when successful:**
```
status: "success"
pipelines: [
  {id: "A", status: "success", result: "..."},
  {id: "B", status: "success", result: "..."},
  {id: "C", status: "success", result: "..."}
]
parallel_metrics: {
  serialized_pairs: 1,
  parallel_executed: 2,
  total_time: <parallel time, not sum>
}
```

**Result when failed:**
```
status: "partial"
pipelines: [
  {id: "A", status: "success", result: "..."},
  {id: "B", status: "failure", reason: "...", rollback: true},
  {id: "C", status: "success", result: "..."}
]
```

**How a model uses this:**
Model says "build backend + build frontend + run integration tests" → orchestrator creates 3 pipelines → checks conflicts → backend and frontend parallel, integration test serialized after both succeed. Model receives results from all three.

---

## Principles From Sources

### embryo (pkg/orchestrator/)

**Principles taken:**
- Pipeline: state→mesh→DIRECT→classify→exec→flywheel→respond. Each stage enriches context.
- Rollback: failed stage receives error context, does not restart from beginning.
- Context flow: cumulative, forward-only. Later phases see everything from earlier phases.

**What Mimic does with them:**
6-phase pipeline with explicit rollback paths. Context accumulates per session.

### bun (PR #30412)

**Principles taken:**
- Phase graph: 6 phases with gate transitions. No skips.
- 2-vote adversarial verify: two independent verifiers for critical operations.
- Edit scope isolation: resource_bitmask per operation prevents cross-contamination.
- Pre/post hooks: PreToolUse/PostToolUse middleware chain.
- Never-rules: hardcoded deny set, no override.

**What Mimic does with them:**
All 6 phases mandatory. Never-rules enforced before every operation. 2-vote on critical ops. Hooks on every tool call.

### gastown

**Principles taken:**
- ZFC state: observable reality over cached assumptions.
- Event-driven convoy: completion event triggers dependency-aware dispatch.
- Pressure gating: system load gates dispatch.
- Help classification: triage incoming requests by severity.

**What Mimic does with them:**
Every phase checks observable reality. Completion events trigger next tasks. Pressure gates prevent overload.

### graphify

**Principles taken:**
- AST extraction: two-pass (structural + call-graph) with confidence labels.
- IDF-weighted search: exact > prefix > substring with gap-ratio cutoff.
- Hub-throttled traversal: skip high-degree hubs as transit.

**What Mimic does with them:**
Pattern retrieval uses 5-signal ranking. High-confidence patterns preferred. Hub nodes skipped in traversal.

---

## Artifact Storage

How orchestrator processes become mesh slots:

| Field | Value |
|-------|-------|
| domain | "orchestrator" |
| layer | "process" |
| modality | "code" |
| pattern_name | "intent_to_validated_chain" / "mesh_query" / "parallel_pipeline_orchestration" |
| pattern_code | semantic chain description |
| invariants | ["no_exec_without_validate", "no_phase_skip", "budget_never_exceeded", "complete_honest_response"] |
| survival_index | from usage across all domains |
| z_density | computed from usage frequency + survival |
| polarity | POSITIVE |

---

## Cross-Domain Conflicts

Orchestrator domain coordinates with all other domains but does not conflict. It is the conductor, not an instrument. However:
- Orchestrator PLAN phase reads from all domains.
- Orchestrator VALIDATE phase checks conflicts across all domains.
- Orchestrator EXEC phase delegates to domain-specific executors.
