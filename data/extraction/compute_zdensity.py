#!/usr/bin/env python3
"""
compute_zdensity.py — Compute Z-density for a seed bmap file.

Z(slot) = (Σ survival_i × weight_i) / slot_volume
From BEHAVIOR.md #5.

Usage: python3 compute_zdensity.py --seed <bmap> --repo <org/repo>
"""

import argparse
import json
import sys


def compute_zdensity(seed_path, repo):
    try:
        with open(seed_path, "r") as f:
            data = json.load(f)
    except Exception:
        return 0.0

    artifacts = data.get("artifacts", [])
    if not artifacts:
        return 0.0

    total_si = sum(a.get("survival_index", 0.0) for a in artifacts)
    total_tokens = sum(max(a.get("token_count", 1), 1) for a in artifacts)

    z = total_si / total_tokens if total_tokens > 0 else 0.0
    return round(z, 6)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--seed", required=True)
    parser.add_argument("--repo", default="")
    args = parser.parse_args()

    z = compute_zdensity(args.seed, args.repo)
    print(f"{z}")


if __name__ == "__main__":
    main()
