#!/usr/bin/env python3
"""
compute_survival_v2.py — Compute survival index from git blame output.

Output format per line: sha total_lines survival_index filepath

Usage: python3 compute_survival_v2.py --blame <file> --output <file> --threshold 0.7
"""

import argparse
import sys
from collections import defaultdict


def parse_blame(blame_path):
    """Parse git blame -l -t output. Returns: commit -> {added, surviving, files(set)}."""
    commit_data = defaultdict(lambda: {"added": 0, "surviving": 0, "files": set()})

    with open(blame_path, "r") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            # git blame -l -t format:
            # ^commit_sha filename (author timestamp timezone lineno) content
            # First token = SHA (with optional ^ prefix)
            parts = line.split(None, 1)
            if not parts:
                continue
            sha = parts[0].lstrip("^")
            if len(sha) != 40:
                continue
            # Extract filename: second token before the (
            rest = parts[1] if len(parts) > 1 else ""
            # Find the ' (' that precedes author info
            paren_idx = rest.find(" (")
            filename = rest[:paren_idx].strip() if paren_idx > 0 else "unknown"
            commit_data[sha]["added"] += 1
            commit_data[sha]["surviving"] += 1
            commit_data[sha]["files"].add(filename)

    return commit_data


def compute_survival_index(commit_data, threshold=0.7):
    results = []
    for sha, data in commit_data.items():
        total = data["added"]
        surviving = data["surviving"]
        si = surviving / total if total > 0 else 0.0
        if si >= threshold or si <= 0.1:
            for filename in data["files"]:
                results.append((sha, total, si, filename))
    return results


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--blame", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--threshold", type=float, default=0.7)
    args = parser.parse_args()

    commit_data = parse_blame(args.blame)
    results = compute_survival_index(commit_data, args.threshold)

    with open(args.output, "w") as f:
        for sha, total, si, filename in results:
            f.write(f"{sha} {total} {si:.4f} {filename}\n")

    print(f"[survival] {len(results)} commit-files written to {args.output}", file=sys.stderr)


if __name__ == "__main__":
    main()
