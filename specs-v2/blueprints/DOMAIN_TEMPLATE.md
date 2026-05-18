# Domain Blueprint — TEMPLATE

Copy this to create a new domain spec: domains/<name>/PROCESS.md

---

# <Domain_Name>

## What This Domain Does

One paragraph: what operations belong to this domain, what the model can accomplish through it.

---

## Processes

### <process_name>

**When to use:**
Describe the intent that triggers this process.

**Goal:**
What must be true after this process completes.

**Chain (semantically):**
Step-by-step tokenized process. Each step = one OpPacket or one internal validation.

**Hard constraints:**
What is absolutely forbidden during this process. Not warnings — hard stops.

**Invariants:**
What must hold before, during, and after execution.

**Result when successful:**
What the model receives back.

**Result when failed:**
What the model receives back, including rollback state.

**How a model uses this:**
Example: "Model says 'commit these files' → Mimic returns this process → model follows step by step."

---

## Principles From Sources

### <source_repo>

**Principles taken:**
- list
- of
- principles

**What Mimic does with them:**
How these principles become behavior in Mimic.

**What Mimic does NOT copy:**
What stays in the source repo, what is reimplemented.

---

## Artifact Storage

How processes from this domain become mesh slots.

| Field | Value for this domain |
|-------|----------------------|
| domain | "<name>" |
| layer | "process" |
| modality | "code" |
| pattern_name | process identifier |
| invariants | list of domain invariants |

---

## Cross-Domain Conflicts

Which domains this domain conflicts with and why.
