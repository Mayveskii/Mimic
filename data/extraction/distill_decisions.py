#!/usr/bin/env python3
"""
distill_decisions.py — Extract decision patterns from PR/issue comments.

Computes decision_survival = months_not_reverted / months_since_decision.
Decision with survival >= 1.0 → candidate for deep cache.

Usage: python3 distill_decisions.py --repo <org>/<repo> [--pr N] [--token TOKEN]

Reproducibility: same repo + same PR range → same output (bitwise-identical).
"""

import argparse
import json
import hashlib
import subprocess
import sys
import os
from datetime import datetime, timezone
from typing import List, Dict, Optional

try:
    import yaml
except ImportError:
    yaml = None


def run_gh(args: List[str], token: Optional[str] = None) -> str:
    env = os.environ.copy()
    if token:
        env["GH_TOKEN"] = token
    result = subprocess.run(
        ["gh"] + args,
        capture_output=True, text=True, env=env
    )
    if result.returncode != 0:
        print(f"gh error: {result.stderr}", file=sys.stderr)
        return "[]"
    return result.stdout


def fetch_pr_comments(repo: str, pr_number: int, token: Optional[str] = None) -> List[Dict]:
    stdout = run_gh(
        ["api", f"repos/{repo}/pulls/{pr_number}/comments",
         "--paginate", "--jq", '.[] | {user: .user.login, body: .body, created_at: .created_at}'],
        token
    )
    comments = []
    for line in stdout.strip().split("\n"):
        line = line.strip()
        if line.startswith("{"):
            try:
                comments.append(json.loads(line))
            except json.JSONDecodeError:
                pass
    return comments


def fetch_issue_comments(repo: str, pr_number: int, token: Optional[str] = None) -> List[Dict]:
    stdout = run_gh(
        ["api", f"repos/{repo}/issues/{pr_number}/comments",
         "--paginate", "--jq", '.[] | {user: .user.login, body: .body, created_at: .created_at}'],
        token
    )
    comments = []
    for line in stdout.strip().split("\n"):
        line = line.strip()
        if line.startswith("{"):
            try:
                comments.append(json.loads(line))
            except json.JSONDecodeError:
                pass
    return comments


def parse_decision(text: str) -> Optional[Dict]:
    """
    Extract decision pattern from comment text.
    Looks for: CHOSEN/REJECTED + BECAUSE + MEASURED patterns.
    Returns None if no decision pattern found.
    """
    text_upper = text.upper()

    decision_type = None
    chosen = None
    because = None
    measured = None

    if "CHOSEN:" in text_upper or "CHOSE" in text_upper or "WENT WITH" in text_upper:
        decision_type = "CHOSEN"
    elif "REJECTED:" in text_upper or "REJECT" in text_upper or "NOT DOING" in text_upper:
        decision_type = "REJECTED"
    elif "WHY" in text_upper and ("BECAUSE" in text_upper or "SINCE" in text_upper):
        decision_type = "CHOSEN"
    else:
        return None

    if "BECAUSE" in text_upper:
        because_start = text_upper.index("BECAUSE") + len("BECAUSE")
        because = text[because_start:].strip().split(".")[0].strip()

    if "MEASURED" in text_upper or "BENCHMARK" in text_upper or "%" in text:
        for token in text.split():
            if "%" in token:
                measured = token
                break

    return {
        "decision_type": decision_type,
        "because": because,
        "measured": measured,
        "raw_text": text[:500]
    }


def compute_decision_survival(repo: str, pr_number: int, pr_date: str, token: Optional[str] = None) -> float:
    """
    Check if the decision from this PR has been reverted.
    decision_survival = months_not_reverted / months_since
    """
    pr_dt = datetime.fromisoformat(pr_date.replace("Z", "+00:00"))
    now = datetime.now(timezone.utc)
    months_since = max((now - pr_dt).days / 30.0, 0.1)

    stdout = run_gh(
        ["api", f"repos/{repo}/pulls",
         "--jq", f'.[] | select(.title | test("revert.*#{pr_number}|revert.*{pr_number}"; "i")) | .merged_at'],
        token
    )

    reverted = any(line.strip() for line in stdout.strip().split("\n") if line.strip())

    if reverted:
        return 0.0
    return months_since / months_since  # not reverted = 1.0 if alive

    return 1.0 if not reverted else 0.0


def extract_decisions(repo: str, pr_range: Optional[str] = None, token: Optional[str] = None) -> List[Dict]:
    decisions = []

    if pr_range:
        prs = [int(x) for x in pr_range.split(",") if x.strip().isdigit()]
    else:
        stdout = run_gh(
            ["api", f"repos/{repo}/pulls",
             "--jq", '.[] | select(.state == "closed" and .merged_at != null) | .number',
             "--paginate"],
            token
        )
        prs = [int(x) for x in stdout.strip().split("\n") if x.strip().isdigit()]

    for pr_num in prs[:50]:
        comments = fetch_pr_comments(repo, pr_num, token)
        comments += fetch_issue_comments(repo, pr_num, token)

        for comment in comments:
            decision = parse_decision(comment.get("body", ""))
            if decision:
                decision["pr_number"] = pr_num
                decision["author"] = comment.get("user", "unknown")
                decision["created_at"] = comment.get("created_at", "")
                decision["decision_survival"] = compute_decision_survival(
                    repo, pr_num, comment.get("created_at", ""), token
                )
                decision["source_repo"] = repo
                decision["domain"] = "llm"

                decision_str = json.dumps(decision, sort_keys=True)
                decision["id"] = hashlib.sha256(decision_str.encode()).hexdigest()[:64]

                decisions.append(decision)

    return decisions


def main():
    parser = argparse.ArgumentParser(description="Extract decision patterns from PR comments")
    parser.add_argument("--repo", required=True, help="org/repo")
    parser.add_argument("--pr", help="Specific PR number or range (1,2,3)")
    parser.add_argument("--token", help="GitHub token")
    parser.add_argument("--output", default="-", help="Output file (- for stdout)")
    parser.add_argument("--format", choices=["yaml", "json"], default="yaml")

    args = parser.parse_args()

    decisions = extract_decisions(args.repo, args.pr, args.token)

    if args.format == "json":
        output = json.dumps(decisions, indent=2, sort_keys=True)
    elif yaml:
        output = yaml.dump({"patterns": decisions}, sort_keys=True, allow_unicode=True)
    else:
        output = json.dumps(decisions, indent=2, sort_keys=True)

    if args.output == "-":
        print(output)
    else:
        with open(args.output, "w") as f:
            f.write(output)


if __name__ == "__main__":
    main()
