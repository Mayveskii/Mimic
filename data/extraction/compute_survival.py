#!/usr/bin/env python3
"""
compute_survival.py — Compute survival index from git blame output.

survival(commit) = surviving_lines / total_lines_added

Threshold: >= 0.7 → slot candidate, < 0.1 → discard.

Usage: python3 compute_survival.py --blame <file> --output <file> --threshold 0.7
"""

import argparse
import sys
from collections import defaultdict


def parse_blame(blame_path):
    """Parse git blame -l -t output: SHA filename (author timestamp timezone lineno) content"""
    commit_lines = defaultdict(lambda: {"added": 0, "surviving": 0})

    with open(blame_path, "r") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            # git blame -l -t format:
            # ^commit_sha filename (author timestamp timezone lineno) content
            # First token = SHA (with optional ^ prefix for boundary commits)
            parts = line.split(None, 1)
            if not parts:
                continue
            sha = parts[0].lstrip("^")
            if len(sha) != 40:
                continue
            commit_lines[sha]["added"] += 1
            commit_lines[sha]["surviving"] += 1

    return commit_lines


def compute_survival_index(commit_lines, threshold=0.7):
    results = []
    for sha, counts in commit_lines.items():
        total = counts["added"]
        surviving = counts["surviving"]
        si = surviving / total if total > 0 else 0.0
        results.append((sha, total, si))

    results.sort(key=lambda x: x[2], reverse=True)
    return [(sha, total, si) for sha, total, si in results if si >= threshold or si <= 0.1]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--blame", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--threshold", type=float, default=0.7)
    args = parser.parse_args()

    commit_lines = parse_blame(args.blame)
    results = compute_survival_index(commit_lines, args.threshold)

    with open(args.output, "w") as f:
        for sha, total, si in results:
            f.write(f"{sha} {total} {si:.4f}\n")

    print(f"[survival] {len(results)} commits written to {args.output}", file=sys.stderr)


if __name__ == "__main__":
    main()
