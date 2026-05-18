#!/usr/bin/env python3
"""
artifact_completeness.py — Validate artifact against specs-v2/artifacts/ARTIFACT_SCHEMA.md required fields.

Usage: python3 artifact_completeness.py --artifact <path>
Returns: JSON report with missing_required, missing_conditional, verdict.
"""

import argparse
import json
import sys


REQUIRED_FIELDS = [
    "artifact_id",
    "version",
    "created_at",
    "polarity",
    "slot",
    "sources",
    "extraction",
]

SLOT_REQUIRED = [
    "slot_id",
    "name",
    "domain",
    "layer",
    "modality",
]

SLOT_QS_REQUIRED = [
    "survival_index",
    "z_density",
]

SOURCE_REQUIRED = [
    "repo",
    "commit",
    "path",
]

EXTRACTION_REQUIRED = [
    "tool",
    "hash",
]

CONDITIONAL_RULES = [
    ("counter_pattern_id", "polarity", lambda v: v == "NEGATIVE"),
    ("anti_pattern_id", "polarity", lambda v: v == "NEGATIVE"),
    ("failure_evidence", "polarity", lambda v: v == "NEGATIVE"),
    ("qac_violated", "polarity", lambda v: v == "NEGATIVE"),
]


def check_required(artifact):
    """Check top-level required fields."""
    missing = []
    for field in REQUIRED_FIELDS:
        if field not in artifact or artifact[field] is None:
            missing.append(field)
    return missing


def check_slot(slot):
    """Check slot sub-fields."""
    missing = []
    for field in SLOT_REQUIRED:
        if field not in slot or slot[field] is None:
            missing.append(f"slot.{field}")
    qs = slot.get("quality_signals", {})
    for field in SLOT_QS_REQUIRED:
        if field not in qs or qs[field] is None:
            missing.append(f"slot.quality_signals.{field}")
    return missing


def check_sources(sources):
    """Check source provenance records."""
    missing = []
    if not sources or len(sources) == 0:
        missing.append("sources (must have >= 1)")
        return missing
    for i, s in enumerate(sources):
        for field in SOURCE_REQUIRED:
            if field not in s or s[field] is None:
                missing.append(f"sources[{i}].{field}")
    return missing


def check_extraction(extraction):
    """Check extraction metadata."""
    missing = []
    for field in EXTRACTION_REQUIRED:
        if field not in extraction or extraction[field] is None:
            missing.append(f"extraction.{field}")
    return missing


def check_conditional(artifact):
    """Check conditional fields based on polarity."""
    missing = []
    for field, trigger_field, trigger_fn in CONDITIONAL_RULES:
        if trigger_fn(artifact.get(trigger_field)):
            if field not in artifact or artifact[field] is None or artifact[field] == "":
                missing.append(f"conditional.{field} (required when polarity={artifact.get(trigger_field)})")
    return missing


def validate(artifact):
    missing = []
    missing.extend(check_required(artifact))
    slot = artifact.get("slot", {})
    missing.extend(check_slot(slot))
    missing.extend(check_sources(artifact.get("sources", [])))
    missing.extend(check_extraction(artifact.get("extraction", {})))
    missing.extend(check_conditional(artifact))

    version = artifact.get("version")
    version_ok = version == 2

    qac = artifact.get("assessment", {}).get("qac_mapping", {})

    verdict = "PASS" if len(missing) == 0 and version_ok else "FAIL"
    if not version_ok:
        missing.append(f"version (expected 2, got {version})")

    return {
        "verdict": verdict,
        "missing_fields": missing,
        "has_version_2": version_ok,
        "qac_present": len(qac) > 0,
        "qac_count": len(qac),
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--artifact", required=True)
    parser.add_argument("--output", default="-")
    args = parser.parse_args()

    with open(args.artifact, "r") as f:
        artifact = json.load(f)

    report = validate(artifact)
    output = json.dumps(report, indent=2)

    if args.output == "-":
        print(output)
    else:
        with open(args.output, "w") as f:
            f.write(output)

    sys.exit(0 if report["verdict"] == "PASS" else 1)


if __name__ == "__main__":
    main()
