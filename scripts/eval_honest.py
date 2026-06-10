#!/usr/bin/env python3
"""
Honest Eval Harness for v16Q.
Collects proof after execution and produces structured eval output.
Rule: no eval without proof.
"""

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path
from typing import Any


def run_cmd(cmd: list[str], cwd: str | None = None) -> str:
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, cwd=cwd, timeout=30)
        return result.stdout.strip()
    except Exception as e:
        return f"ERROR: {e}"


def collect_proof(worktree_path: str) -> dict[str, Any]:
    return {
        "git_status": run_cmd(["git", "-C", worktree_path, "status", "--porcelain"]),
        "git_log": run_cmd(["git", "-C", worktree_path, "log", "--oneline", "-5"]),
        "ls_la": run_cmd(["ls", "-la", worktree_path]),
        "git_diff": run_cmd(["git", "-C", worktree_path, "diff", "--stat"]),
    }


def eval_run(worktree_path: str, model_id: str, intent: str) -> dict[str, Any]:
    proof = collect_proof(worktree_path)
    return {
        "model_id": model_id,
        "intent": intent,
        "worktree_path": worktree_path,
        "proof": proof,
        "timestamp": int(os.times().elapsed),
    }


def main():
    parser = argparse.ArgumentParser(description="v16Q Honest Eval Harness")
    parser.add_argument("--worktree", required=True, help="Path to worktree")
    parser.add_argument("--model", required=True, help="Model ID")
    parser.add_argument("--intent", required=True, help="Intent string")
    parser.add_argument("--output", default="eval.json", help="Output JSON path")
    args = parser.parse_args()

    if not os.path.isdir(os.path.join(args.worktree, ".git")):
        print(f"ERROR: {args.worktree} is not a git worktree", file=sys.stderr)
        sys.exit(1)

    result = eval_run(args.worktree, args.model, args.intent)

    out_path = Path(args.output)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with open(out_path, "w") as f:
        json.dump(result, f, indent=2)

    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
