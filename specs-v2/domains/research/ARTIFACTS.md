# Research Domain — Artifacts

Mesh slots produced by the research domain.

---

## Slot Types

| Type | layer | Purpose | Linked To |
|------|-------|---------|-----------|
| HYPOTHESIS | "hypothesis" | Stated research question + prediction | EXPERIMENT slots |
| EXPERIMENT | "experiment" | Parameterized measurement design | RESULT slots |
| RESULT | "result" | Raw data + statistical analysis | HYPOTHESIS slot |
| LITERATURE | "literature" | Parsed paper with metadata | CITATION slots |
| INFERENCE | "inference" | Conclusion from result + hypothesis | HYPOTHESIS + RESULT |
| PROGRESS | "progress" | Cross-session key findings | None (standalone) |

---

## Hypothesis Slot Structure

```json
{
  "slot_type": "RESEARCH",
  "layer": "hypothesis",
  "domain": "research",
  "payload": {
    "id": "hyp-uuid",
    "statement": "X causes Y in context Z",
    "predicted_outcome": "measurement M will exceed threshold T",
    "experiment_design_id": "exp-uuid",
    "status": "proposed | running | confirmed | refuted | refined",
    "parent_hypothesis_id": null,
    "refinement_of": null
  }
}
```

---

## Experiment Slot Structure

```json
{
  "slot_type": "RESEARCH",
  "layer": "experiment",
  "domain": "research",
  "payload": {
    "id": "exp-uuid",
    "hypothesis_id": "hyp-uuid",
    "parameters": {
      "tool_calls": [...],
      "data_sources": [...],
      "control_variables": [...]
    },
    "reproducibility_hash": "sha256(params+tools+versions)"
  }
}
```

---

## Result Slot Structure

```json
{
  "slot_type": "RESEARCH",
  "layer": "result",
  "domain": "research",
  "payload": {
    "id": "res-uuid",
    "hypothesis_id": "hyp-uuid",
    "experiment_id": "exp-uuid",
    "raw_data": "...|large_data_next_slot_id",
    "statistical_test": {
      "test_name": "t-test | chi-sq | anova | ...",
      "p_value": 0.042,
      "ci_lower": 0.12,
      "ci_upper": 0.34,
      "effect_size": 0.23,
      "sample_size": 100
    },
    "timestamp": "ISO-8601"
  }
}
```

---

## Literature Slot Structure

```json
{
  "slot_type": "RESEARCH",
  "layer": "literature",
  "domain": "research",
  "payload": {
    "id": "lit-uuid",
    "doi": "10.xxxx/xxxxx",
    "url": "https://arxiv.org/abs/...",
    "title": "...",
    "abstract": "...",
    "authors": [...],
    "year": 2024,
    "citations": ["lit-uuid-2", "lit-uuid-3"],
    "cited_by": ["lit-uuid-4"],
    "embedding_id": "emb-uuid"
  }
}
```

Literature text > 64KB is split with `next_slot_id` chain.

---

## Progress Slot Structure

```json
{
  "slot_type": "RESEARCH",
  "layer": "progress",
  "domain": "research",
  "payload": {
    "session_id": "sess-uuid",
    "hypotheses_updated": ["hyp-uuid-1", "hyp-uuid-2"],
    "key_findings": ["Finding A: ...", "Finding B: ..."],
    "next_steps": ["Run experiment X", "Search literature on Y"],
    "context_summary": "semantic compressed text"
  }
}
```

---

## OpPacket Chains for Research

### Create hypothesis and run experiment

```
[0] OP_RESEARCH_HYPOTHESIS_CREATE    {statement: "...", predicted: "..."}
[1] OP_RESEARCH_EXPERIMENT_RUN       {hypothesis_id: "...", params: {...}}
[2] OP_RESEARCH_RESULT_STORE         {experiment_id: "...", data: "..."}
[3] OP_RESEARCH_STATISTICAL_TEST     {result_id: "...", test: "t-test"}
[4] OP_RESEARCH_HYPOTHESIS_INFERENCE {hypothesis_id: "...", result_id: "..."}
```

### Ingest literature

```
[0] OP_RESEARCH_LITERATURE_FETCH     {doi: "10.xxxx/xxxxx"}
[1] OP_RESEARCH_LITERATURE_PARSE     {}
[2] OP_RESEARCH_LITERATURE_INDEX     {parsed_id: "..."}
[3] OP_RESEARCH_CITATION_LINK        {slot_id: "...", cited_slots: [...]}
[4] OP_RESEARCH_LITERATURE_EMBED     {slot_id: "..."}
```
