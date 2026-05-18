#!/usr/bin/env python3
"""
parse_pr.py — Parse PR review artifacts into structured decision patterns.

Extracts:
  - CHOSEN/REJECTED decisions from PR descriptions and review comments
  - Before/after measurements
  - Rejected alternatives with reasons
  - Correctness claims (bitwise-identical, test suite, etc.)

Follows the pattern from gonka-ai/vllm#36:
  Summary → Why → Measured impact → Correctness → Alternatives not done → Test plan

Usage: python3 parse_pr.py --input <pr_data.json> --output <decisions.json>
"""

import argparse
import json
import re
import hashlib
from typing import List, Dict, Optional, Tuple


SECTION_PATTERNS = {
    "summary": r"(?:##\s*Summary|##\s*Description|Summary)\s*\n(.*?)(?=\n##|\Z)",
    "why": r"(?:##\s*Why|Why\s+(?:it\s+)?(?:helps|this\s+works))\s*\n(.*?)(?=\n##|\Z)",
    "measured": r"(?:##\s*Measured|##\s*(?:Impact|Results|Performance))\s*\n(.*?)(?=\n##|\Z)",
    "correctness": r"(?:##\s*Correctness|##\s*Verification)\s*\n(.*?)(?=\n##|\Z)",
    "alternatives": r"(?:##\s*(?:Things\s+considered|Alternatives|Not\s+done))\s*\n(.*?)(?=\n##|\Z)",
    "test_plan": r"(?:##\s*Test\s+plan|##\s*Testing)\s*\n(.*?)(?=\n##|\Z)",
}


MEASUREMENT_REGEX = re.compile(
    r"(?P<config>[\w\s,=×]+?)\s*"
    r"(?P<before>\d[\d,]*)\s*→\s*(?P<after>\d[\d,]*)\s*"
    r"(?P<unit>\w+(?:/\w+)?)\s*"
    r"(?:\((?P<delta>[+\-]\d+(?:\.\d+)?%?)\))?"
)

PERCENT_DELTA = re.compile(r"[+\-]?\d+(?:\.\d+)?%")

REJECTION_PATTERNS = re.compile(
    r"(?:rejected|not\s+doing|was\s+tested\s+and\s+rejected|decided\s+against|ruled\s+out)"
    r"\s*[:\-]?\s*(.+?)(?:\n|$)",
    re.IGNORECASE
)

REASON_PATTERNS = re.compile(
    r"(?:because|due\s+to|caused?\s+by|reason:)\s*(.+?)(?:\n|$)",
    re.IGNORECASE
)

CORRECTNESS_PATTERNS = re.compile(
    r"(?:bitwise.idential|semantics.preserving|output\s+matches|test\s+suite\s+passes?|no\s+regression)",
    re.IGNORECASE
)


def extract_sections(text: str) -> Dict[str, str]:
    sections = {}
    for name, pattern in SECTION_PATTERNS.items():
        match = re.search(pattern, text, re.DOTALL | re.IGNORECASE)
        if match:
            sections[name] = match.group(1).strip()
    return sections


def extract_measurements(text: str) -> List[Dict]:
    results = []
    for match in MEASUREMENT_REGEX.finditer(text):
        before_str = match.group("before").replace(",", "")
        after_str = match.group("after").replace(",", "")
        try:
            before = int(before_str)
            after = int(after_str)
        except ValueError:
            continue

        delta_str = match.group("delta") or ""
        if delta_str:
            delta_pct = PERCENT_DELTA.search(delta_str)
            if delta_pct:
                delta = float(delta_pct.group().rstrip("%"))
            else:
                delta = round(((after - before) / before) * 100, 1) if before > 0 else 0.0
        else:
            delta = round(((after - before) / before) * 100, 1) if before > 0 else 0.0

        results.append({
            "config": match.group("config").strip(),
            "before": before,
            "after": after,
            "unit": match.group("unit"),
            "delta_pct": delta
        })
    return results


def extract_rejections(text: str) -> List[Dict]:
    rejections = []
    for match in REJECTION_PATTERNS.finditer(text):
        alt = match.group(1).strip()
        reason_match = REASON_PATTERNS.search(text[match.end():])
        reason = reason_match.group(1).strip() if reason_match else "not specified"
        rejections.append({"alternative": alt, "reason": reason})
    return rejections


def extract_correctness_claims(text: str) -> List[str]:
    claims = []
    for match in CORRECTNESS_PATTERNS.finditer(text):
        claims.append(match.group(0).lower())
    return claims


def parse_pr(pr_data: Dict) -> Dict:
    body = pr_data.get("body", "")
    title = pr_data.get("title", "")
    full_text = f"# Summary\n{title}\n\n{body}"

    sections = extract_sections(full_text)
    measurements = extract_measurements(sections.get("measured", body))
    rejections = extract_rejections(sections.get("alternatives", body))
    correctness = extract_correctness_claims(full_text)

    decision_type = "CHOSEN"
    if rejections and not measurements:
        decision_type = "REJECTED"

    main_decision = sections.get("why", sections.get("summary", title))

    artifact = {
        "pattern_name": re.sub(r"[^a-z0-9]+", "_", title.lower())[:64],
        "decision": f"{decision_type}: {main_decision[:200]}",
        "measurements": measurements,
        "rejections": rejections,
        "correctness_claims": correctness,
        "pr_number": pr_data.get("number", 0),
        "source_repo": pr_data.get("head", {}).get("repo", {}).get("full_name", ""),
        "invariants": []
    }

    for claim in correctness:
        artifact["invariants"].append({
            "condition": claim,
            "source": f"PR#{pr_data.get('number', 0)}",
            "verification": "CI"
        })

    artifact_str = json.dumps(artifact, sort_keys=True)
    artifact["id"] = hashlib.sha256(artifact_str.encode()).hexdigest()[:64]

    return artifact


def main():
    parser = argparse.ArgumentParser(description="Parse PR artifacts into decision patterns")
    parser.add_argument("--input", required=True, help="PR data JSON file")
    parser.add_argument("--output", default="-", help="Output file")
    args = parser.parse_args()

    with open(args.input, "r") as f:
        pr_data = json.load(f)

    if isinstance(pr_data, list):
        artifacts = [parse_pr(pr) for pr in pr_data]
    else:
        artifacts = [parse_pr(pr_data)]

    output = json.dumps(artifacts, indent=2, sort_keys=True)

    if args.output == "-":
        print(output)
    else:
        with open(args.output, "w") as f:
            f.write(output)


if __name__ == "__main__":
    main()
