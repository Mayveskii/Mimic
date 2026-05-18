#!/usr/bin/env python3
"""
extract_patterns.py — Extract code patterns from filtered survival-index file.

Reads surviving commit lines, pulls out corresponding code from the repo,
splits into function-level patterns, enforces max token limit.

Usage: python3 extract_patterns.py --filtered <file> --repo-dir <path> --output <file>
"""

import argparse
import sys
from pathlib import Path


def extract_patterns(filtered_path, repo_dir, max_tokens=2048):
    patterns = []
    seen = set()

    with open(filtered_path, "r") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            # compute_survival_v2.py format: sha total_lines survival_index filename
            # filename may contain spaces, so split into 4+ parts:
            parts = line.split(" ", 3)
            if len(parts) < 4:
                continue

            sha = parts[0]
            total_str = parts[1]
            si_str = parts[2]
            rel_path = parts[3]  # filename/path, may have spaces

            key = f"{sha}:{rel_path}"
            if key in seen:
                continue
            seen.add(key)

            full_path = Path(repo_dir) / rel_path
            if not full_path.exists():
                continue

            try:
                content = full_path.read_text(errors="replace")
            except Exception:
                continue

            token_count = len(content) // 4
            if token_count > max_tokens:
                chunks = chunk_code(content, max_tokens)
                for i, chunk in enumerate(chunks):
                    pattern_id = f"{sha}:{rel_path}:chunk{i}"
                    patterns.append(f"{pattern_id}\t{rel_path}\t{len(chunk)//4}\t{chunk[:max_tokens*4]}")
            else:
                pattern_id = f"{sha}:{rel_path}:full"
                patterns.append(f"{pattern_id}\t{rel_path}\t{token_count}\t{content[:max_tokens*4]}")

    return patterns


def chunk_code(content, max_tokens):
    max_chars = max_tokens * 4
    lines = content.split("\n")
    chunks = []
    current = []

    for line in lines:
        current.append(line)
        if len("\n".join(current)) > max_chars:
            chunks.append("\n".join(current[:-1]))
            current = [line]

    if current:
        chunks.append("\n".join(current))

    return chunks


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--filtered", required=True)
    parser.add_argument("--repo-dir", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--max-tokens", type=int, default=2048)
    args = parser.parse_args()

    patterns = extract_patterns(args.filtered, args.repo_dir, args.max_tokens)

    with open(args.output, "w") as f:
        for p in patterns:
            f.write(p + "\n")

    print(f"[patterns] {len(patterns)} patterns written to {args.output}", file=sys.stderr)


if __name__ == "__main__":
    main()
