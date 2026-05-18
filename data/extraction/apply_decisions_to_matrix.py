#!/usr/bin/env python3
"""
apply_decisions_to_matrix.py — Parse decision-patterns.yaml and emit C code for conflict_matrix + energy_cost entries.

Usage: python3 apply_decisions_to_matrix.py --input <yaml> --output-c <c-file> --output-h <h-file>
"""

import argparse
import json
import sys
from pathlib import Path


def parse_matrix_impact(pattern):
    """Extract matrix_impact from a decision pattern entry."""
    impact = pattern.get("matrix_impact", {})
    conflict_rules = impact.get("conflict_rules_added", [])
    energy_costs = impact.get("energy_costs_added", [])
    domains = impact.get("domains_expanded", [])
    return conflict_rules, energy_costs, domains


def generate_c_code(patterns):
    header_lines = [
        "/* Auto-generated from decision-patterns.yaml by apply_decisions_to_matrix.py */",
        "#ifndef MIMIC_DECISION_MATRIX_H",
        "#define MIMIC_DECISION_MATRIX_H",
        "",
        "#include \"ops.h\"",
        "",
        "/* Decision-derived conflict rules */",
        "typedef struct {",
        "    OpCode op1;",
        "    OpCode op2;",
        "    ConflictLevel level;",
        "    const char* source;",
        "    const char* rationale;",
        "} DecisionConflictRule;",
        "",
        "/* Decision-derived energy costs */",
        "typedef struct {",
        "    OpCode opcode;",
        "    float cost_tokens;",
        "    float cost_time_us;",
        "    bool measured;",
        "    const char* source;",
        "} DecisionEnergyCost;",
        "",
        "void ops_load_decision_conflicts(void);",
        "void ops_load_decision_energy(void);",
        "",
        "#endif",
    ]

    c_lines = [
        "/* Auto-generated from decision-patterns.yaml by apply_decisions_to_matrix.py */",
        '#include "decision_matrix.h"',
        "",
        "static DecisionConflictRule g_decision_conflicts[] = {",
    ]

    conflict_count = 0
    for pattern in patterns:
        rules, _, _ = parse_matrix_impact(pattern)
        for rule in rules:
            conflict_count += 1
            c_lines.append(f'    {{ {rule["op1"]}, {rule["op2"]}, {rule["level"]}, "{pattern["source_repo"]}", "{rule.get("reason", "")}" }},')

    c_lines.append("};")
    c_lines.append(f"")
    c_lines.append(f"static const size_t g_decision_conflict_count = {conflict_count};")
    c_lines.append("")
    c_lines.append("static DecisionEnergyCost g_decision_energy[] = {")

    energy_count = 0
    for pattern in patterns:
        _, costs, _ = parse_matrix_impact(pattern)
        for cost in costs:
            energy_count += 1
            measured_str = "true" if cost.get("measured", False) else "false"
            c_lines.append(f'    {{ {cost["opcode"]}, {cost["cost_tokens"]}f, {cost["cost_time_us"]}f, {measured_str}, "{cost.get("source", "")}" }},')

    c_lines.append("};")
    c_lines.append(f"")
    c_lines.append(f"static const size_t g_decision_energy_count = {energy_count};")
    c_lines.append("")
    c_lines.append("void ops_load_decision_conflicts(void) {")
    c_lines.append("    for (size_t i = 0; i < g_decision_conflict_count; i++) {")
    c_lines.append("        DecisionConflictRule* r = &g_decision_conflicts[i];")
    c_lines.append("        g_conflict_matrix[r->op1][r->op2] = r->level;")
    c_lines.append("        g_conflict_matrix[r->op2][r->op1] = r->level;")
    c_lines.append("    }")
    c_lines.append("}")
    c_lines.append("")
    c_lines.append("void ops_load_decision_energy(void) {")
    c_lines.append("    for (size_t i = 0; i < g_decision_energy_count; i++) {")
    c_lines.append("        DecisionEnergyCost* c = &g_decision_energy[i];")
    c_lines.append("        g_energy_costs[c->opcode][0] = c->cost_tokens;")
    c_lines.append("        g_energy_costs[c->opcode][1] = c->cost_time_us;")
    c_lines.append("    }")
    c_lines.append("}")

    return "\n".join(header_lines) + "\n", "\n".join(c_lines) + "\n"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", default="mimicrya/decision-patterns.yaml")
    parser.add_argument("--output-c", default="data/matrices/decision_matrix.c")
    parser.add_argument("--output-h", default="data/matrices/decision_matrix.h")
    args = parser.parse_args()

    try:
        import yaml
        with open(args.input, "r") as f:
            data = yaml.safe_load(f)
    except ImportError:
        # Fallback: parse as JSON if yaml not available
        with open(args.input, "r") as f:
            data = json.load(f)

    patterns = data.get("patterns", [])
    header, code = generate_c_code(patterns)

    Path(args.output_c).parent.mkdir(parents=True, exist_ok=True)
    with open(args.output_c, "w") as f:
        f.write(code)
    with open(args.output_h, "w") as f:
        f.write(header)

    print(f"[matrix] Generated {args.output_c} and {args.output_h} from {len(patterns)} patterns")


if __name__ == "__main__":
    main()
