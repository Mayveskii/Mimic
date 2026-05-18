# Anti-Patterns Domain — Sources

Where the anti-patterns domain behavior comes from.

---

## binary-mesh History

**Principles taken:**
- 30 documented anti-patterns (AP-01..AP-30) with evidence, counter_patterns, QAC mapping.
- 5 new domains: resource_cleanup_under_lock, context_structure_stability, input_validation_before_io, idempotent_close_cleanup, economic_invariant_enforcement.
- Meta-invariant: all 30 trace to missing validation.

**What Mimic does with them:**
All 30 APs documented with evidence and counters. QAC mapping complete. New domains integrated.

---

## gonka-ai/vllm (PR #36)

**Principles taken:**
- Evidence-based performance claims: measured before/after with specific hardware.
- Bitwise-identical verification for correctness.
- Rejected alternatives documented with reasons.
- Weight diff changes as evidence.

**What Mimic does with them:**
Anti-patterns and patterns both require evidence. Performance claims backed by measurements. Alternatives evaluated and rejected with justification.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- pkg/checkpoints/: C6-C10 gates include anti-pattern detection.
- pkg/hunt/: hunt pipeline flags potential anti-patterns.

**What Mimic does with them:**
Anti-patterns detected at checkpoints. Hunt pipeline flags during distillation.

---

## gastown

**Principles taken:**
- ZFC state: anti-patterns are observable failures, not assumptions.
- Rollback: anti-patterns often discovered during failed rollback.

**What Mimic does with them:**
Anti-patterns documented from real failures. Rollback failures add to evidence.

---

## Standard Practice

**Principles taken:**
- Root cause analysis: 5 whys.
- Evidence-based documentation.
- Counter-examples with solutions.
- Regular review and archival.

**What Mimic does with them:**
Standard quality assurance practices applied to anti-pattern management.
