# Research Domain — Process Specification

How Mimic supports long-running scientific research: hypothesis tracking, experiment design, literature ingestion, result collation, and multi-session continuity.

---

## Core Processes

### 1. Hypothesis Lifecycle

```
Hypothesis: stated intent with falsifiable prediction
    |
    v
[OP_RESEARCH_HYPOTHESIS_CREATE]  → creates hypothesis artifact with ID
    |
    v
Experiment: design + execute + measure
    |
    v
Result: observed data vs predicted
    |
    v
Inference: confirm / refute / refine
    |
    v
New hypothesis (if refuted) → loop
```

**Behavior**: When agent expresses a research goal, Mimic creates a hypothesis artifact. The artifact stores:
- The hypothesis statement (falsifiable)
- Predicted outcome
- Experiment design (tool calls, data sources)
- Actual results (measured)
- Inference (confirm/refute/refine)

**Result**: Every research step is tracked. No result is lost. 

**Why**: Research requires traceability. Without hypothesis tracking, agent drifts between approaches without learning.

---

### 2. Experiment Design

```
Intent: "verify that X causes Y"
    |
    v
PLAN:
  [0] OP_RESEARCH_LITERATURE_SEARCH  → find existing work on X→Y
  [1] OP_RESEARCH_DATA_SOURCE        → identify dataset/tool for measurement
  [2] OP_RESEARCH_EXPERIMENT_RUN     → execute measurement with recorded params
  [3] OP_RESEARCH_RESULT_STORE       → store raw result with hash
  [4] OP_RESEARCH_STATISTICAL_TEST   → compute significance / confidence interval
```

**Behavior**: Experiment is an OpPacket chain like any other. Each step is validated before execution. Results stored in mesh slots with survival-weighted indexing.

**Result**: Raw data + statistical analysis stored as retrievable artifacts.

**Why**: Scientific results must be reproducible. Every parameter, every dataset, every tool version must be recorded.

---

### 3. Literature Ingestion

```
Input: PDF, arXiv URL, DOI, bibtex entry
    |
    v
[0] OP_RESEARCH_LITERATURE_FETCH    → download/fetch content
[1] OP_RESEARCH_LITERATURE_PARSE    → extract: title, abstract, methods, results, citations
[2] OP_RESEARCH_LITERATURE_INDEX    → create mesh slot: domain="research", layer="literature"
[3] OP_RESEARCH_CITATION_LINK       → link cited papers (slot_id → slot_id)
[4] OP_RESEARCH_LITERATURE_EMBED    → generate embedding for semantic search
```

**Behavior**: Literature is treated as mesh slots. Citations form a graph. Full text chunked into linked slots if > 64KB.

**Result**: Searchable knowledge base of papers. Exact text retrievable.

**Why**: Research builds on prior work. Without literature indexing, agent reinvents or misses key findings.

---

### 4. Multi-Session Research Continuity

```
Session N ends:
  [0] OP_SESS_SNAPSHOT                → full session state compressed
  [1] OP_RESEARCH_PROGRESS_STORE      → key findings as high-signal mesh slots
  [2] OP_RESEARCH_CONTEXT_SUMMARIZE   → semantic compression of session context

Session N+1 begins:
  [0] OP_SESS_RESTORE                 → load prior session state
  [1] OP_RESEARCH_HYPOTHESIS_LOAD     → reload active hypotheses
  [2] OP_RAG_QUERY (session_history)  → "what did I conclude last time?"
```

**Behavior**: Research spans sessions. Context is never lost. Agent resumes with full knowledge of prior work.

**Result**: Seamless multi-day research without context loss.

**Why**: Complex research cannot complete in one session. Checkpoint + resume + history RAG is essential.

---

### 5. Tool Chaining for Research

```
Tool output → next tool input:
  OP_RESEARCH_LITERATURE_SEARCH results → feed into OP_RESEARCH_EXPERIMENT_RUN params
  OP_RESEARCH_DATA_SOURCE findings → feed into OP_BUILD_COMPILE target
  OP_BUILD_TEST results → feed into OP_RESEARCH_STATISTICAL_TEST
  OP_RESEARCH_STATISTICAL_TEST p-value → feed into OP_RESEARCH_HYPOTHESIS_INFERENCE
```

**Behavior**: Tool outputs are structured (JSON). Next tool inputs reference prior output fields. Validation checks type compatibility.

**Result**: Composable research workflows. Output of search becomes input of experiment.

**Why**: Research is not linear. Data from one tool drives the next. Wiring outputs to inputs enables automation.

---

## Domain-Specific OpCodes

| OpCode | Purpose | Safety |
|--------|---------|--------|
| OP_RESEARCH_HYPOTHESIS_CREATE | Create a falsifiable hypothesis | SAFE |
| OP_RESEARCH_HYPOTHESIS_LOAD | Load existing hypothesis by ID | READONLY |
| OP_RESEARCH_HYPOTHESIS_INFERENCE | Confirm/refute/refine based on results | SAFE |
| OP_RESEARCH_EXPERIMENT_RUN | Execute experiment with recorded params | DANGEROUS* |
| OP_RESEARCH_RESULT_STORE | Store raw result with hash | DANGEROUS |
| OP_RESEARCH_STATISTICAL_TEST | Compute significance / CI | SAFE |
| OP_RESEARCH_LITERATURE_FETCH | Download/fetch paper | SAFE |
| OP_RESEARCH_LITERATURE_PARSE | Extract structured info | SAFE |
| OP_RESEARCH_LITERATURE_INDEX | Index as mesh slot | DANGEROUS |
| OP_RESEARCH_CITATION_LINK | Link cited papers | SAFE |
| OP_RESEARCH_LITERATURE_EMBED | Generate embedding | SAFE |
| OP_RESEARCH_PROGRESS_STORE | Store key findings for cross-session | DANGEROUS |
| OP_RESEARCH_CONTEXT_SUMMARIZE | Semantic compression of session | SAFE |

*DANGEROUS = requires explicit allow. See domain/security/PROCESS.md.

---

## Invariants

- HYP-INV-1: Every hypothesis MUST be falsifiable (have a predicted outcome that can be proven false)
- HYP-INV-2: Every experiment result MUST link to the hypothesis it tests
- HYP-INV-3: Every literature slot MUST contain DOI or URL for verification
- HYP-INV-4: Research progress MUST be stored as mesh slots before session end
- HYP-INV-5: Statistical test results MUST include confidence interval, not just p-value
- HYP-INV-6: Tool chain outputs MUST be typed (schema-validated before next tool input)

---

## Sources

| Behavior | Source | Evidence |
|----------|--------|----------|
| Hypothesis tracking | scientific method/peer review | universal research practice |
| Experiment reproducibility | terraform state file pattern | hashicorp/terraform plan/apply cycle |
| Literature indexing | arXiv API + bibtex parsing | academic workflow |
| Multi-session continuity | gastown session_continuity | Mayveskii/gastown |
| Tool chaining | langchain chain abstractions | langchain-ai/langchain |
| Statistical testing | scipy.stats / R patterns | scientific computing standards |

---

## Artifacts

Research artifacts stored in mesh slots:

```
slot_type: RESEARCH
layer: "hypothesis" | "experiment" | "result" | "literature" | "inference"
domain: "research"
payload: {<structured data>}
linked_slots: [<citation slot IDs>]
```
