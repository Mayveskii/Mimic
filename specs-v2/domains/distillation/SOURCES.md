# Distillation Domain — Sources

Where the distillation domain behavior comes from.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- pkg/survival/: git blame → survival_index = surviving_lines / total_lines.
- pkg/hunt/: 17-module hunt pipeline includes distillation assessment.
- pkg/checkpoints/: C6-C10 gates include distillation checkpoints.

**What Mimic does with them:**
Survival index computed from git blame. Hunt pipeline assesses distillation quality. Checkpoints gate progression.

**What Mimic does NOT copy:**
- Embryo's specific blame parsing logic.
- Embryo's Go-based pipeline.

---

## gonka-ai/vllm (PR #36)

**Principles taken:**
- Performance optimization with measured impact: compile hot path, measure before/after.
- Bitwise-identical verification: output of optimized vs baseline must match exactly.
- A/B testing with pristine extraction.
- Evidence-based claims: tracker links showing weight diff changes.

**What Mimic does with them:**
Distillation extracts patterns with measurable impact. Every pattern includes baseline and optimized metrics. Survival index correlates with measured improvement.

**What Mimic does NOT copy:**
- vLLM's torch.compile usage.
- vLLM's specific GPU optimization.

---

## graphify

**Principles taken:**
- AST extraction: two-pass structural + call-graph.
- Confidence labels: every extraction has confidence score.
- IDF-weighted relevance.

**What Mimic does with them:**
Pattern extraction uses structural analysis. Confidence scores feed into artifact_precision.

---

## bun (PR #30412)

**Principles taken:**
- Phase graph: structured pipeline with gates.
- Quality assessment: every stage has validation.

**What Mimic does with them:**
Distillation pipeline follows structured phases. Quality gates at each stage.

---

## gastown

**Principles taken:**
- ZFC state: observable reality over assumptions.
- Event-driven: completion triggers next stage.

**What Mimic does with them:**
Distillation verifies source repo state before extraction. Completion events trigger indexing.

---

## Standard Practice

**Principles taken:**
- Git blame for provenance.
- Reproducible builds (hash everything).
- Quality thresholds before acceptance.
- Evidence-based anti-patterns.

**What Mimic does with them:**
Standard software engineering practices applied to pattern extraction.
