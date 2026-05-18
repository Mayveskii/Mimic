# Artifact Schema — Distillation Output Structure

An artifact is the output of the distillation pipeline. It links a proven pattern (slot) with its extraction provenance, quality assessment, and mesh placement metadata.

Artifacts are JSON-serialized for pipeline stages, then converted to binary slots for mesh storage.

---

## JSON Artifact Structure

```json
{
  "artifact_id": "art-uuid-v4",
  "version": 2,
  "created_at": "2026-05-17T12:00:00Z",
  
  "polarity": "POSITIVE",
  "counter_pattern_id": null,
  "anti_pattern_id": null,
  
  "slot": {
    "slot_id": 18446744073709551615,
    "name": "rollback_on_failure",
    "domain": "resource_cleanup",
    "layer": "process",
    "modality": "code",
    
    "quality_signals": {
      "survival_index": 0.92,
      "z_density": 0.81,
      "artifact_precision": 0.95,
      "usage_frequency": 0.05
    },
    
    "invariants": [
      "partial_failure_leaves_no_orphans",
      "every_alloc_has_matching_free"
    ],
    
    "tags": [
      "error-handling",
      "resource-cleanup",
      "golang",
      "defer-pattern"
    ]
  },
  
  "sources": [
    {
      "repo": "github.com/gastown/polecat",
      "commit": "a1b2c3d4e5f6...",
      "path": "internal/polecat/manager.go",
      "line_start": 142,
      "line_end": 189,
      "blame_timestamp": 1715923456,
      "extraction_confidence": 0.94
    }
  ],
  
  "extraction": {
    "tool": "distill-v2.1",
    "parameters": {
      "language": "go",
      "granularity": "function",
      "depth": 3
    },
    "hash": "sha256:abc123...",
    "duration_ms": 450
  },
  
  "failure_evidence": null,
  "qac_violated": [],
  
  "assessment": {
    "qac_mapping": {
      "QAC-1": "pass",
      "QAC-2": "pass",
      "QAC-7": "pass"
    },
    "reviewer_notes": ""
  }
}
```

---

## Required Fields

| Field | Type | Required | Description |
|---|---|---|---|
| artifact_id | UUID string | yes | Unique artifact identifier |
| version | integer | yes | Schema version (currently 2) |
| created_at | ISO-8601 | yes | Creation timestamp |
| polarity | enum string | yes | POSITIVE / NEGATIVE / COUNTER |
| slot | object | yes | Core pattern data |
| sources | array | yes | ≥1 source provenance record |
| extraction | object | yes | Tool and parameter metadata |

---

## Conditional Fields

| Field | When Required | Description |
|---|---|---|
| counter_pattern_id | polarity == NEGATIVE | UUID of the COUNTER artifact |
| anti_pattern_id | polarity == NEGATIVE | ID of anti-pattern record (AP-NN) |
| failure_evidence | polarity == NEGATIVE | Description of observed failure |
| qac_violated | polarity == NEGATIVE | List of QAC codes violated by this anti-pattern |

---

## Polarity Rules

1. **POSITIVE**: Documents a correct pattern. `counter_pattern_id` and `anti_pattern_id` MUST be null.
2. **NEGATIVE**: Documents an anti-pattern (what went wrong). MUST have:
   - `counter_pattern_id`: UUID of the correct alternative.
   - `anti_pattern_id`: Reference to anti-pattern record (e.g., "AP-05").
   - `failure_evidence`: Concrete description of failure (not abstract).
   - `qac_violated`: Array of QAC codes this anti-pattern violates.
3. **COUNTER**: The correct alternative to a NEGATIVE. `counter_pattern_id` points back to the NEGATIVE it addresses.

---

## Source Provenance Record

```c
typedef struct {
    char repo[128];              // Repository URL or local path
    char commit[40];             // Full hex SHA-1
    char path[256];              // File path
    uint32_t line_start;         // First line of extracted pattern
    uint32_t line_end;           // Last line of extracted pattern
    uint64_t blame_timestamp;    // Git blame timestamp (author date)
    float extraction_confidence; // 0.0-1.0, tool confidence score
} SourceRecord;
```

Rules:
- `blame_timestamp` is the author date of the commit that introduced these lines (git blame -t).
- `extraction_confidence` is the tool's confidence that this region contains a complete, meaningful pattern.
- Multiple sources (up to 4) indicate the pattern was independently discovered in multiple repos.

---

## Extraction Metadata

```c
typedef struct {
    char tool[32];               // Extractor name + version
    char parameters_hash[64];    // SHA-256 of serialized parameters
    uint64_t duration_ms;        // Extraction time
    char hash[64];               // SHA-256 of extractor binary + parameters + source content
} ExtractionMetadata;
```

The `hash` field ensures reproducibility: given the same extractor, parameters, and source content, the hash must match. If hash mismatches on re-extraction → source content changed or tool version differs.

---

## QAC Mapping

```c
typedef struct {
    char qac_code[8];            // "QAC-NN" format
    char result[8];              // "pass", "fail", "na"
    char notes[256];             // Optional reviewer notes
} QACAssessment;
```

Every artifact must assess against all 13 Quality Assurance Criteria:
- QAC-1 through QAC-13.
- Result is "pass", "fail", or "na" (not applicable).
- "fail" on any QAC → artifact_precision < 1.0.
- More than 3 "fail" → artifact_precision < 0.8 → slot NOT indexed.

---

## Conversion to Slot

Artifact → Slot conversion is lossy (JSON to binary):
- `artifact_id` is not stored in slot (slot uses `slot_id`).
- JSON fields are flattened to binary struct fields.
- Sources are truncated to SLOT_MAX_SOURCES (4).
- Invariants are truncated to SLOT_MAX_INVARIANTS (16).
- Tags are truncated to SLOT_MAX_TAGS (32).
- Text content is stored at `text_offset` with `text_len`.

Reverse conversion (Slot → Artifact) is used for API responses:
- Slot ID mapped to artifact UUID via external registry.
- Binary fields expanded to JSON.
- Retrieval statistics added.

---

## Artifact Lifecycle

```
EXTRACTION
    ↓ tool runs on source code
VALIDATION
    ↓ QAC assessment, precision check
DECISION
    ↓ precision ≥ 0.8? → CONVERT TO SLOT → INDEX
    ↓ precision < 0.8? → ARCHIVE (not indexed, logged)
FEEDBACK LOOP
    ↓ After application, success/failure recorded
    ↓ If failure_count / retrieval_count > 0.2 → FLAG FOR REVIEW
    ↓ If survival_index drops (source repo deleted) → ARCHIVE
```
