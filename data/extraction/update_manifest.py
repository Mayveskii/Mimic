#!/usr/bin/env python3
"""
update_manifest.py — Update repos-manifest.yaml after distillation.

Sets status, last_commit, slots, z_density for a specific repo.

Usage: python3 update_manifest.py --manifest <path> --repo <org/repo> --commit <sha> --slots N --z-density X --status distilled
"""

import argparse
import sys
from pathlib import Path


def update_manifest(manifest_path, repo, commit, slots, z_density, status):
    text = Path(manifest_path).read_text()

    lines = text.split("\n")
    in_repo = False
    updated = False
    indent = "  "

    new_lines = []
    i = 0
    while i < len(lines):
        line = lines[i]

        if f"- repo: {repo}" in line:
            in_repo = True
            new_lines.append(line)
            i += 1
            continue

        if in_repo:
            if line.strip().startswith("status:"):
                new_lines.append(f"{indent}status: {status}")
                updated = True
            elif line.strip().startswith("last_commit:"):
                new_lines.append(f"{indent}last_commit: \"{commit}\"")
                updated = True
            elif line.strip().startswith("slots:"):
                new_lines.append(f"{indent}slots: {slots}")
                updated = True
            elif line.strip().startswith("z_density:"):
                new_lines.append(f"{indent}z_density: {z_density}")
                updated = True
            elif line.strip().startswith("- repo:") or (not line.startswith(indent) and line.strip()):
                in_repo = False
                new_lines.append(line)
            else:
                new_lines.append(line)
        else:
            new_lines.append(line)

        i += 1

    Path(manifest_path).write_text("\n".join(new_lines))

    if updated:
        print(f"[manifest] updated {repo}: status={status} slots={slots} z={z_density}", file=sys.stderr)
    else:
        print(f"[manifest] WARNING: {repo} not found in manifest", file=sys.stderr)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", required=True)
    parser.add_argument("--repo", required=True)
    parser.add_argument("--commit", required=True)
    parser.add_argument("--slots", type=int, required=True)
    parser.add_argument("--z-density", required=True)
    parser.add_argument("--status", default="distilled")
    args = parser.parse_args()

    update_manifest(args.manifest, args.repo, args.commit, args.slots, args.z_density, args.status)


if __name__ == "__main__":
    main()
