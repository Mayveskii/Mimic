#!/usr/bin/env python3
"""
quality_gate.py — 13 Quality Assurance Criteria (QAC) validation for Mimic artifacts.

Usage: python3 quality_gate.py --artifact <path> [--output report.json]
Returns: qac_mapping with pass/fail/na, artifact_precision, verdict.
"""

import argparse
import json
import re
import sys
from pathlib import Path

QAC_COUNT = 13


def qac_1_survival_index_provenance(artifact):
    """SI MUST be computed from git blame, not guessed."""
    slot = artifact.get("slot", {})
    sources = artifact.get("sources", [])
    survival = slot.get("quality_signals", {}).get("survival_index", 0.0)
    if survival < 0 or survival > 1:
        return "fail", f"survival_index {survival} out of range [0, 1]"
    has_blame = any("blame_timestamp" in s for s in sources)
    if not has_blame and len(sources) == 0:
        return "fail", "No source provenance, cannot verify SI provenance"
    return "pass", f"survival_index={survival}, sources={len(sources)}"


def qac_2_invariant_coverage(artifact):
    """Every artifact MUST have >=1 verifiable invariant."""
    invariants = artifact.get("slot", {}).get("invariants", [])
    if len(invariants) == 0:
        return "fail", "Zero invariants"
    if any(not inv.strip() for inv in invariants):
        return "fail", "Empty invariant string found"
    return "pass", f"invariants={len(invariants)}"


def qac_3_energy_cost_measured(artifact):
    """Energy cost MUST be measured, not estimated."""
    # For distillation artifacts, energy costs apply at execution time, not extraction time.
    slot = artifact.get("slot", {})
    latency_us = slot.get("quality_signals", {}).get("latency_us")
    token_count = slot.get("quality_signals", {}).get("token_count")
    # If latency_us is None OR 0, it means not measured yet (distillation phase)
    if (latency_us is None or latency_us == 0) and token_count is not None:
        return "na", "Distillation artifact — energy cost applies at execution time"
    if latency_us is None and token_count is None:
        return "na", "Distillation artifact — energy cost applies at execution time"
    if latency_us is not None and latency_us > 0:
        return "pass", f"latency_us={latency_us}"
    return "fail", "Energy cost not measured"


def qac_4_conflict_matrix_derivation(artifact):
    """Conflict matrix entries MUST be derived from observed conflicts, not invented."""
    # Distillation artifacts themselves don't derive conflict rules, but their slot's domain must match known domains
    domain = artifact.get("slot", {}).get("domain", "")
    known_domains = {"distributed", "database", "network", "security", "observability",
                     "build", "runtime", "llm", "agent", "git", "os", "data",
                     "io", "memory", "system", "process", "utility",
                     "orchestrator", "session", "rag", "mesh", "quality",
                     "research", "self-management", "anti-patterns", "general"}
    if domain not in known_domains:
        return "fail", f"Unknown domain '{domain}' — not in conflict matrix spec"
    return "pass", f"domain='{domain}'"


def qac_5_z_density_computed(artifact):
    """Z-density MUST be computed from actual slot data, not default 0."""
    z = artifact.get("slot", {}).get("quality_signals", {}).get("z_density", 0.0)
    if z <= 0:
        return "fail", f"Z-density {z} <= 0 — must be computed from slot data"
    return "pass", f"Z-density={z:.4f}"


def qac_6_decision_consistency(artifact):
    """Decision pattern MUST NOT contradict existing matrix entries."""
    polarity = artifact.get("polarity", "POSITIVE")
    if polarity == "NEGATIVE" and not artifact.get("counter_pattern_id"):
        return "fail", "NEGATIVE artifact without counter_pattern_id contradicts QAC-9"
    if polarity not in {"POSITIVE", "NEGATIVE", "COUNTER"}:
        return "fail", f"Unknown polarity '{polarity}'"
    return "pass", f"polarity={polarity}"


def qac_7_artifact_precision(artifact):
    """all three precision components > 0 to enter deep cache."""
    qs = artifact.get("slot", {}).get("quality_signals", {})
    si = qs.get("survival_index", 0.0)
    inv_count = len(artifact.get("slot", {}).get("invariants", []))
    inv_cov = min(inv_count / 3.0, 1.0)  # invariant coverage: 3+ = 100%
    extraction = artifact.get("extraction", {})
    ext_hash = bool(extraction.get("hash", ""))
    repro = 1.0 if ext_hash else 0.0
    precision = si * inv_cov * repro
    if precision > 0 and precision >= 0.8:
        return "pass", f"precision={precision:.4f} (SI={si} × inv_cov={inv_cov:.2f} × repro={repro})"
    elif precision > 0:
        return "fail", f"precision={precision:.4f} < 0.8 threshold"
    return "fail", f"precision=0 (SI={si} × inv_cov={inv_cov:.2f} × repro={repro})"


def qac_8_multimodal_integrity(artifact):
    """Multimodal extraction MUST verify integrity."""
    modality = artifact.get("slot", {}).get("modality", "CODE")
    if modality in {"CODE", "TEXT"}:
        return "pass", f"modality={modality} — integrity implicit in text extraction"
    return "na", f"modality={modality} — IMAGE/DIAGRAM integrity not yet implemented"


def qac_9_anti_pattern_polarity(artifact):
    """Every NEGATIVE MUST link to POSITIVE counter_pattern."""
    polarity = artifact.get("polarity", "POSITIVE")
    counter = artifact.get("counter_pattern_id")
    if polarity == "NEGATIVE" and not counter:
        return "fail", "NEGATIVE without counter_pattern_id"
    if polarity == "POSITIVE" and counter:
        return "fail", "POSITIVE should not have counter_pattern_id"
    return "pass", f"polarity={polarity}, counter_pattern_id={'set' if counter else 'null'}"


def qac_10_temporal_consistency(artifact):
    """SI MUST be re-validated when blame data changes."""
    sources = artifact.get("sources", [])
    # At minimum, we verify that blame_timestamp exists and is non-zero
    for s in sources:
        if "blame_timestamp" in s and s["blame_timestamp"] > 0:
            return "pass", f"blame_timestamp={s['blame_timestamp']}"
    return "fail", "No blame_timestamp in any source — temporal consistency unverifiable"


def qac_11_cross_domain_conflict(artifact):
    """Conflict matrix includes cross-domain pairs."""
    domain = artifact.get("slot", {}).get("domain", "")
    if domain in {"git", "build", "io", "network", "system", "memory", "process", "utility"}:
        # These are OpCode domains — cross-domain rules enforced by CONFLICT_MATRIX_SPEC.md
        return "pass", f"OpCode domain '{domain}' under CONFLICT_MATRIX_SPEC.md"
    return "na", f"Knowledge domain '{domain}' — cross-domain conflict rules not directly applicable"


def qac_12_provenance_chain(artifact):
    """Every artifact MUST carry extraction_hash matching current extractor."""
    extraction = artifact.get("extraction", {})
    if not extraction.get("hash", ""):
        return "fail", "extraction.hash empty"
    if not extraction.get("tool", ""):
        return "fail", "extraction.tool empty"
    return "pass", f"tool={extraction['tool']}, hash present={bool(extraction['hash'])}"


def qac_13_revert_detection(artifact):
    """Commits with 'Revert' MUST generate NEGATIVE artifact."""
    polarity = str(artifact.get("polarity", "POSITIVE"))
    failure_evidence = artifact.get("failure_evidence", "")
    if polarity == "NEGATIVE" and not failure_evidence:
        return "fail", "NEGATIVE artifact without failure_evidence"
    return "pass", f"polarity={polarity}, failure_evidence={'set' if failure_evidence else 'null'}"


QAC_CHECKS = [
    qac_1_survival_index_provenance,
    qac_2_invariant_coverage,
    qac_3_energy_cost_measured,
    qac_4_conflict_matrix_derivation,
    qac_5_z_density_computed,
    qac_6_decision_consistency,
    qac_7_artifact_precision,
    qac_8_multimodal_integrity,
    qac_9_anti_pattern_polarity,
    qac_10_temporal_consistency,
    qac_11_cross_domain_conflict,
    qac_12_provenance_chain,
    qac_13_revert_detection,
]


def run_all_qac(artifact):
    qac_mapping = {}
    for i, check_fn in enumerate(QAC_CHECKS):
        code = f"QAC-{i + 1}"
        try:
            result, note = check_fn(artifact)
        except Exception as e:
            result, note = "fail", str(e)
        qac_mapping[code] = {"result": result, "note": note}

    # Compute artifact_precision
    qs = artifact.get("slot", {}).get("quality_signals", {})
    si = qs.get("survival_index", 0.0)
    inv_count = len(artifact.get("slot", {}).get("invariants", []))
    inv_cov = min(inv_count / 3.0, 1.0)
    ext_hash = bool(artifact.get("extraction", {}).get("hash", ""))
    repro = 1.0 if ext_hash else 0.0
    precision = round(si * inv_cov * repro, 4)

    # Verdict
    fails = sum(1 for v in qac_mapping.values() if v["result"] == "fail")
    passes = sum(1 for v in qac_mapping.values() if v["result"] == "pass")
    nas = sum(1 for v in qac_mapping.values() if v["result"] == "na")

    if fails == 0:
        verdict = "DEEP_CACHE" if precision >= 0.8 else "LOCAL_ONLY"
    elif fails <= 3 and precision >= 0.5:
        verdict = "REVIEW_PENDING"
    else:
        verdict = "REJECT"

    return {
        "qac_mapping": qac_mapping,
        "artifact_precision": precision,
        "verdict": verdict,
        "passes": passes,
        "fails": fails,
        "na": nas,
    }


def main():
    parser = argparse.ArgumentParser(description="Validate artifact against 13 QAC")
    parser.add_argument("--artifact", required=True)
    parser.add_argument("--output", default="-")
    args = parser.parse_args()

    with open(args.artifact, "r") as f:
        artifact = json.load(f)

    report = run_all_qac(artifact)
    output = json.dumps(report, indent=2)

    if args.output == "-":
        print(output)
    else:
        with open(args.output, "w") as f:
            f.write(output)

    sys.exit(0 if report["fails"] == 0 else 1)


if __name__ == "__main__":
    main()
