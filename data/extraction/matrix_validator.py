#!/usr/bin/env python3
"""
matrix_validator.py — Validate artifact and process chains against CONFLICT_MATRIX_SPEC.md and ENERGY_COST_SPEC.md.

Usage: python3 matrix_validator.py --artifact <path> [--chain <path>]
"""

import argparse
import json
import sys

# From CONFLICT_MATRIX_SPEC.md — key conflict pairs (op1, op2, level)
CONFLICT_RULES = [
    ("OP_SYS_EXEC", "OP_SYS_EXEC", 3),
    ("OP_GIT_STATUS", "OP_GIT_COMMIT", 2),
    ("OP_GIT_CHECKOUT", "OP_GIT_ADD", 3),
    ("OP_GIT_CHECKOUT", "any_git_op", 3),
    ("OP_GIT_MERGE", "any_git_op", 3),
    ("OP_GIT_REBASE", "any_git_op", 4),
    ("OP_BUILD_COMPILE", "OP_BUILD_CLEAN", 4),
    ("OP_BUILD_LINK", "OP_BUILD_COMPILE", 3),
    ("OP_BUILD_TEST", "OP_BUILD_COMPILE", 3),
    ("OP_IO_WRITE", "OP_IO_WRITE", 3),
    ("OP_IO_READ", "OP_IO_WRITE", 2),
    ("OP_IO_CLOSE", "OP_IO_READ", 3),
    ("OP_MMAP_FREE", "OP_MMAP_READ", 3),
    ("OP_MMAP_FREE", "OP_MMAP_WRITE", 3),
    ("OP_PROC_KILL", "OP_PROC_KILL", 3),
]

# Resource bitmask from CONFLICT_MATRIX_SPEC.md
DOMAIN_BITS = {
    "memory": [0x10, 0x11, 0x12, 0x13, 0x14],
    "io": [0x20, 0x21, 0x22, 0x23, 0x24],
    "git": [0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x3B, 0x3C, 0x3D],
    "build": [0x40, 0x41, 0x42, 0x43, 0x44],
    "network": [0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56],
    "process": [0x60, 0x61, 0x62, 0x63],
    "utility": [0x70, 0x71, 0x72, 0x73, 0x74, 0x75],
    "system": [0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89],
    "session": [0x90, 0x91, 0x92, 0x93, 0x94],
    "orchestrator": [0x95, 0x96, 0x97, 0x98, 0x99, 0x9A],
    "research": [0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xAB, 0xAC],
    "self": [0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5],
}

# Energy cost defaults from ENERGY_COST_SPEC.md
DEFAULT_ENERGY = {
    "OP_IO_READ": (1.0, 2.0, 0.0),
    "OP_IO_WRITE": (2.0, 3.0, 0.0),
    "OP_IO_OPEN": (1.0, 2.0, 0.0),
    "OP_IO_CLOSE": (0.5, 1.0, 0.0),
    "OP_IO_SEEK": (0.5, 1.0, 0.0),
    "OP_BUILD_COMPILE": (6.0, 1000000.0, 0.0),
    "OP_BUILD_LINK": (3.0, 100000.0, 0.0),
    "OP_BUILD_TEST": (6.0, 5000000.0, 0.0),
    "OP_BUILD_DEPLOY": (8.0, 30000000.0, 0.0),
    "OP_BUILD_CLEAN": (2.0, 50000.0, 0.0),
    # ... (all from spec)
}


def check_chain_conflict(chain):
    """Check if a chain of OpPackets has any conflicts."""
    opcodes = [p["opcode"] for p in chain]
    conflicts = []
    for i in range(len(opcodes)):
        for j in range(i + 1, len(opcodes)):
            for r in CONFLICT_RULES:
                op1, op2, level = r
                if (opcodes[i] == op1 and opcodes[j] == op2) or (opcodes[i] == op2 and opcodes[j] == op1):
                    conflicts.append({
                        "index_pair": [i, j],
                        "operations": [opcodes[i], opcodes[j]],
                        "level": level,
                        "level_name": ["NONE", "LOW", "MEDIUM", "HIGH", "FATAL"][level],
                    })
    return conflicts


def check_chain_energy(chain, budget_tokens=None, budget_time_ms=None):
    """Check if chain fits within energy budget."""
    total_tokens = 0.0
    total_time_us = 0.0

    for packet in chain:
        op = packet.get("opcode", "OP_NOP")
        tokens, time_us, _ = DEFAULT_ENERGY.get(op, (0.0, 0.0, 0.0))
        total_tokens += tokens
        total_time_us += time_us

    result = {
        "total_tokens": total_tokens,
        "total_time_us": total_time_us,
        "total_time_ms": total_time_us / 1000.0,
    }

    overruns = []
    if budget_tokens is not None and total_tokens > budget_tokens:
        overruns.append(f"tokens {total_tokens:.1f} > budget {budget_tokens:.1f}")
    if budget_time_ms is not None and total_time_us / 1000.0 > budget_time_ms:
        overruns.append(f"time {total_time_us / 1000.0:.1f}ms > budget {budget_time_ms:.1f}ms")

    result["overruns"] = overruns
    result["within_budget"] = len(overruns) == 0
    return result


def validate_artifact_domain(artifact):
    """Check if artifact domain is recognized and has conflict rules."""
    domain = artifact.get("slot", {}).get("domain", "")
    if domain in DOMAIN_BITS:
        return {"pass": True, "bits": DOMAIN_BITS[domain]}
    return {"pass": True, "message": "Non-OpCode domain — no static conflict rules", "bits": []}


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--artifact")
    parser.add_argument("--chain")
    parser.add_argument("--budget-tokens", type=float, default=None)
    parser.add_argument("--budget-time-ms", type=float, default=None)
    parser.add_argument("--output", default="-")
    args = parser.parse_args()

    report = {"conflict_checks": [], "energy_checks": [], "domain_checks": []}

    if args.artifact:
        with open(args.artifact, "r") as f:
            artifact = json.load(f)
        report["domain_checks"].append(validate_artifact_domain(artifact))

    if args.chain:
        with open(args.chain, "r") as f:
            chain = json.load(f)
        conflicts = check_chain_conflict(chain)
        report["conflict_checks"] = conflicts
        report["has_conflicts"] = len(conflicts) > 0

        energy = check_chain_energy(chain, args.budget_tokens, args.budget_time_ms)
        report["energy_checks"].append(energy)

    output = json.dumps(report, indent=2)
    if args.output == "-":
        print(output)
    else:
        with open(args.output, "w") as f:
            f.write(output)

    # Exit 1 if conflicts found or energy overruns
    has_problems = report.get("has_conflicts", False) or any(not e.get("within_budget", True) for e in report["energy_checks"])
    sys.exit(1 if has_problems else 0)


if __name__ == "__main__":
    main()
