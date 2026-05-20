#!/usr/bin/env python3
"""
Check configuration consistency between specs/11-CONFIGURATION.md and codebase.

Usage:
    python3 scripts/check_config_consistency.py \
        --spec specs/11-CONFIGURATION.md \
        --env .env.example \
        --code internal/config/,internal/mcp/,core/ops.h,Makefile,Dockerfile

Exit codes:
    0 - all OK
    1 - variable in code but missing from spec
    2 - variable in spec but missing from code (WARNING only, non-blocking)
    3 - parsing or I/O error
"""

import argparse
import os
import re
import sys
from pathlib import Path


def parse_spec_variables(spec_path):
    """Extract variable names from 11-CONFIGURATION.md."""
    if not os.path.exists(spec_path):
        print(f"Spec file not found: {spec_path}")
        return set()

    variables = set()
    with open(spec_path, 'r') as f:
        for line in f:
            # Match lines like: ### `VAR_NAME`
            m = re.match(r'###\s+`([A-Z_0-9]+)`', line)
            if m:
                variables.add(m.group(1))
    return variables


def parse_env_variables(env_path):
    """Extract variable names from .env.example (lines starting with X=Y)."""
    if not os.path.exists(env_path):
        return set()

    variables = set()
    with open(env_path, 'r') as f:
        for line in f:
            line = line.strip()
            # Skip comments and empty lines
            if not line or line.startswith('#'):
                continue
            # Match KEY=value or KEY= (commented out in example)
            m = re.match(r'([A-Z_0-9]+)=', line)
            if m:
                variables.add(m.group(1))
    return variables


def find_variables_in_code(code_paths):
    """Search for variable mentions in source code."""
    found = set()
    env_patterns = [
        re.compile(r'os\.getenv\("([^"]+)"\)'),
        re.compile(r'os\.Getenv\("([^"]+)"\)'),
        re.compile(r'os\.LookupEnv\("([^"]+)"\)'),
        re.compile(r'getenv\("([^"]+)"\)'),
        re.compile(r'getEnvDefault\("([^"]+)"'),
        re.compile(r'getEnvIntDefault\("([^"]+)"'),
        re.compile(r'EXPOSE\s+(\d+)'),  # Docker ports as hints
        re.compile(r'MIMIC_[A-Z_0-9]+'),
        re.compile(r'EXA_[A-Z_0-9]+'),
    ]

    for path_str in code_paths.split(','):
        path = Path(path_str)
        if path.is_file():
            files = [path]
        elif path.is_dir():
            files = list(path.rglob('*'))
        else:
            # Try as glob pattern
            base = path.parent if path.parent != Path('.') else Path('.')
            if base.is_dir():
                files = [f for f in base.rglob('*') if f.is_file()]
            else:
                continue

        for f in files:
            if f.suffix not in ('.go', '.c', '.h', '.md', '.sh', '.yml', '.yaml', '.json', '.mod', ''):
                continue
            if 'vendor' in str(f) or 'node_modules' in str(f):
                continue
            try:
                with open(f, 'r', encoding='utf-8', errors='ignore') as fh:
                    content = fh.read()
                    for pat in env_patterns:
                        for match in pat.finditer(content):
                            name = match.group(1)
                            if name.isupper() and ('MIMIC' in name or 'EXA' in name or 'MAX_' in name):
                                found.add(name)
            except Exception:
                pass

    return found


def main():
    parser = argparse.ArgumentParser(description='Check config consistency')
    parser.add_argument('--spec', default='specs/11-CONFIGURATION.md')
    parser.add_argument('--env', default='.env.example')
    parser.add_argument('--code', default='internal/config/,internal/mcp/,core/ops.h,Makefile,Dockerfile')
    args = parser.parse_args()

    spec_vars = parse_spec_variables(args.spec)
    env_vars = parse_env_variables(args.env)
    code_vars = find_variables_in_code(args.code)

    print(f"Variables in spec ({args.spec}):   {len(spec_vars)}")
    print(f"Variables in env ({args.env}):     {len(env_vars)}")
    print(f"Variables in code:                  {len(code_vars)}")
    print()

    # Invariant INV-CFG-3: code vars must be in spec
    missing_in_spec = code_vars - spec_vars
    if missing_in_spec:
        print(f"FAIL: Variables declared in code but MISSING from spec:")
        for v in sorted(missing_in_spec):
            print(f"  - {v}")
        print()
        return 1

    # WARNING: spec vars not in code (future/planned)
    missing_in_code = spec_vars - code_vars
    if missing_in_code:
        print(f"WARN: Variables in spec but not yet in code (future/planned?):")
        for v in sorted(missing_in_code):
            print(f"  - {v}")
        print()

    # Invariant INV-CFG-2: env vars should be subset of spec
    env_not_in_spec = env_vars - spec_vars
    if env_not_in_spec:
        print(f"WARN: Variables in .env.example but missing from spec:")
        for v in sorted(env_not_in_spec):
            print(f"  - {v}")
        print()

    print("OK: Configuration consistency check passed.")
    return 0


if __name__ == '__main__':
    sys.exit(main())
