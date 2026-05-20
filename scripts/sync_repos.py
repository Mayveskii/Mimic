#!/usr/bin/env python3
"""
Sync repository list: update repos-manifest.yaml with latest SHAs from upstream.

Usage:
    python3 scripts/sync_repos.py --manifest mimicrya/repos-manifest.yaml --output mimicrya/repos-manifest.yaml
"""

import argparse
import subprocess
import sys
import tempfile
from pathlib import Path

import yaml


def get_latest_sha(repo):
    """Get latest commit SHA from GitHub API (no auth needed for public repos)."""
    import urllib.request
    import json
    api_url = f"https://api.github.com/repos/{repo}/commits/HEAD"
    try:
        req = urllib.request.Request(api_url, headers={'Accept': 'application/vnd.github.v3+json'})
        with urllib.request.urlopen(req, timeout=15) as resp:
            data = json.loads(resp.read().decode('utf-8'))
            return data.get('sha', '')
    except Exception as e:
        print(f"WARN: Could not fetch SHA for {repo}: {e}")
        return ''


def update_repo_shas(manifest):
    """Iterate manifest and update last_commit for pending repos."""
    updated = False
    for key in ('current', 'new', 'recommended', 'agentic', 'git', 'network', 'python', 'llm', 'swe', 'os', 'security', 'hardware'):
        section = manifest.get(key, [])
        if isinstance(section, list):
            for item in section:
                if isinstance(item, dict) and item.get('status') == 'pending':
                    repo = item.get('repo', '')
                    current_sha = item.get('last_commit', '')
                    latest = get_latest_sha(repo)
                    if latest and latest != current_sha:
                        item['last_commit'] = latest
                        updated = True
                        print(f"Updated {repo}: {current_sha[:8]} → {latest[:8]}")
    return updated


def main():
    parser = argparse.ArgumentParser(description='Sync repo SHAs')
    parser.add_argument('--manifest', default='mimicrya/repos-manifest.yaml')
    parser.add_argument('--output', default='mimicrya/repos-manifest.yaml')
    args = parser.parse_args()

    with open(args.manifest) as f:
        manifest = yaml.safe_load(f)

    updated = update_repo_shas(manifest)
    if updated:
        with open(args.output, 'w') as f:
            yaml.dump(manifest, f, default_flow_style=False, allow_unicode=True, sort_keys=False)
        print(f"Wrote updated manifest to {args.output}")
    else:
        print("No changes needed.")

    return 0


if __name__ == '__main__':
    sys.exit(main())
