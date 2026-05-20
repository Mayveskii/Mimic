#!/usr/bin/env python3
"""
Reach — Auto-ingestion via Exa API for pending repos.

Fetches web content for repos in repos-manifest.yaml with status=pending
and 0 slots, then writes structured seeds to data/seeds/.

Usage:
    python3 scripts/reach.py --manifest mimicrya/repos-manifest.yaml --seeds data/seeds --max-results 5
"""

import argparse
import json
import os
import re
import sys
import time
import urllib.request
from pathlib import Path
from datetime import datetime


def load_manifest(manifest_path):
    """Parse YAML-like repos-manifest."""
    import yaml
    with open(manifest_path) as f:
        return yaml.safe_load(f)


def get_pending_repos(manifest):
    """Extract repos with status=pending and slots=0."""
    pending = []
    # The manifest has nested lists under categories
    for key in ('current', 'new', 'recommended', 'agentic', 'git', 'network', 'python', 'llm', 'swe', 'os', 'security', 'hardware'):
        section = manifest.get(key, [])
        if isinstance(section, list):
            for item in section:
                if isinstance(item, dict) and item.get('status') == 'pending' and item.get('slots', 0) == 0:
                    pending.append(item)
        elif isinstance(section, dict):
            for subkey, items in section.items():
                if isinstance(items, list):
                    for item in items:
                        if isinstance(item, dict) and item.get('status') == 'pending' and item.get('slots', 0) == 0:
                            pending.append(item)
    return pending


def exa_search(query, api_key, max_results=5):
    """Call Exa /search endpoint."""
    url = os.getenv('EXA_BASE_URL', 'https://api.exa.ai') + '/search'
    data = json.dumps({
        'query': query,
        'type': 'auto',
        'numResults': max_results,
        'useAutoprompt': True,
    }).encode('utf-8')

    req = urllib.request.Request(
        url,
        data=data,
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {api_key}',
        },
        method='POST'
    )

    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read().decode('utf-8'))
    except urllib.error.HTTPError as e:
        if e.code == 429:
            print(f"WARN: Rate limited on Exa search. Backing off...")
            time.sleep(5)
        print(f"WARN: Exa search error: {e}")
        return {'results': []}
    except Exception as e:
        print(f"WARN: Exa search failed: {e}")
        return {'results': []}


def exa_fetch(urls, api_key, max_chars=2000):
    """Call Exa /contents endpoint."""
    endpoint = os.getenv('EXA_BASE_URL', 'https://api.exa.ai') + '/contents'
    data = json.dumps({
        'urls': urls,
        'text': {'maxCharacters': max_chars, 'includeHtml': False},
    }).encode('utf-8')

    req = urllib.request.Request(
        endpoint,
        data=data,
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {api_key}',
        },
        method='POST'
    )

    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read().decode('utf-8'))
    except Exception as e:
        print(f"WARN: Exa fetch failed: {e}")
        return {'results': []}


def sanitize_filename(name):
    """Make repo name safe for filesystem."""
    return re.sub(r'[^a-zA-Z0-9_-]', '_', name)


def write_seed(seeds_dir, repo_name, query, search_results, fetched_content):
    """Write structured seed JSON."""
    seeds_dir = Path(seeds_dir)
    seeds_dir.mkdir(parents=True, exist_ok=True)

    timestamp = datetime.utcnow().isoformat() + 'Z'
    safe_name = sanitize_filename(repo_name)
    filename = seeds_dir / f"reach_{safe_name}_{int(time.time())}.json"

    seed = {
        'source': 'reach',
        'repo': repo_name,
        'query': query,
        'fetched_at': timestamp,
        'search_results': [
            {
                'title': r.get('title', ''),
                'url': r.get('url', ''),
                'id': r.get('id', ''),
            }
            for r in search_results
        ],
        'fetched_content': fetched_content,
    }

    with open(filename, 'w') as f:
        json.dump(seed, f, indent=2, ensure_ascii=False)

    print(f"  Wrote seed: {filename}")
    return filename


def main():
    parser = argparse.ArgumentParser(description='Reach: auto-ingestion via Exa')
    parser.add_argument('--manifest', default='mimicrya/repos-manifest.yaml')
    parser.add_argument('--seeds', default='data/seeds')
    parser.add_argument('--max-results', type=int, default=5)
    parser.add_argument('--dry-run', action='store_true')
    args = parser.parse_args()

    api_key = os.getenv('EXA_API_KEY')
    if not api_key:
        print("ERROR: EXA_API_KEY not set. Skipping reach.")
        return 1

    print(f"Loading manifest: {args.manifest}")
    try:
        manifest = load_manifest(args.manifest)
    except Exception as e:
        print(f"ERROR: Failed to parse manifest: {e}")
        return 1

    pending = get_pending_repos(manifest)
    print(f"Found {len(pending)} pending repos with 0 slots.")

    if args.dry_run:
        print("DRY RUN: no API calls made.")
        for r in pending:
            print(f"  Would research: {r.get('repo', 'unknown')} — {r.get('notes', '')}")
        return 0

    processed = 0
    for repo_info in pending[:args.max_results]:
        repo = repo_info.get('repo', '')
        notes = repo_info.get('notes', '')
        if not repo:
            continue

        query = f"github {repo} {notes}".strip()
        print(f"\nResearching: {repo}")
        print(f"  Query: {query}")

        search_resp = exa_search(query, api_key, args.max_results)
        results = search_resp.get('results', [])
        print(f"  Found {len(results)} search results.")

        if not results:
            continue

        urls = [r['url'] for r in results if r.get('url')]
        fetched = []
        if urls:
            fetch_resp = exa_fetch(urls[:3], api_key, max_chars=2000)
            for fr in fetch_resp.get('results', []):
                fetched.append({
                    'url': fr.get('url', ''),
                    'title': fr.get('title', ''),
                    'text': fr.get('text', '')[:1000],  # truncate for seed
                })

        write_seed(args.seeds, repo, query, results, fetched)
        processed += 1

        # Rate limit politeness
        time.sleep(1)

    print(f"\nDone. Processed {processed} repos. Seeds written to {args.seeds}/")
    return 0


if __name__ == '__main__':
    sys.exit(main())
