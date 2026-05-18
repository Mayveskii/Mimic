# Research Domain — Invariants

Every invariant in the research domain.

---

## HYP-INV-1: Falsifiability

Every hypothesis MUST have a predicted outcome that can be proven false.

```
validation: predicted_outcome is not empty AND predicted_outcome contains measurable criteria
violation: hypothesis lacks measurable criteria → agent cannot determine confirm/refute
```

---

## HYP-INV-2: Result Traceability

Every experiment result MUST link to the hypothesis it tests.

```
validation: result.hypothesis_id == active_hypothesis.id
violation: orphaned result → cannot evaluate hypothesis
```

---

## HYP-INV-3: Literature Verifiability

Every literature slot MUST contain DOI or URL for verification.

```
validation: slot.doi is not empty OR slot.url is not empty
violation: unverifiable source → agent may cite non-existent work
```

---

## HYP-INV-4: Session Continuity

Research progress MUST be stored as mesh slots before session end.

```
validation: session end triggers OP_RESEARCH_PROGRESS_STORE + OP_SESS_SNAPSHOT
violation: progress lost between sessions → research must restart
```

---

## HYP-INV-5: Statistical Rigor

Statistical test results MUST include confidence interval, not just p-value.

```
validation: result.ci_lower is not empty AND result.ci_upper is not empty
violation: p-value alone is misleading (ASA statement on p-values, 2016)
```

---

## HYP-INV-6: Typed Tool Outputs

Tool chain outputs MUST be schema-validated before next tool input.

```
validation: output_schema matches next_input_schema (structural compatibility check)
violation: type mismatch → tool receives invalid args → chain fails
```

---

## HYP-INV-7: Citation Graph Integrity

Citation links MUST form a DAG (no cycles).

```
validation: depth-first search from any slot finds no back-edge to already-visited slot
violation: circular citation → invalid graph → infinite loops in retrieval
```

---

## HYP-INV-8: Literature Chunking

Literature text > 64KB MUST be split into linked slots.

```
validation: slot.text_size <= MSHINV-05 limit (64KB) OR slot.next_slot_id is set
violation: oversized slot → mesh storage failure
```

---

## HYP-INV-9: Experiment Isolation

Experiments MUST not mutate shared state without prior snapshot.

```
validation: OP_RESEARCH_EXPERIMENT_RUN with state_mutation=true requires snapshot_id
violation: unrecoverable state corruption
```
