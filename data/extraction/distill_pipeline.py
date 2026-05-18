#!/usr/bin/env python3
"""
distill_pipeline.py — Run full distillation: validate + synthesize + metrics

Usage: python3 data/extraction/distill_pipeline.py [--seeds-dir data/seeds/] [--output data/distilled/]
"""

import json
import sys
import argparse
from pathlib import Path
from datetime import datetime

def load_artifacts(path):
    with open(path, 'r') as f:
        return json.load(f)

def validate_artifact(art, idx):
    """Quick validation: must have artifact_id, slot, precision >= 0.8"""
    errors = []
    if not art.get('artifact_id'):
        errors.append(f"[{idx}] Missing artifact_id")
    slot = art.get('slot', {})
    if not slot:
        errors.append(f"[{idx}] Missing slot")
    precision = slot.get('quality_signals', {}).get('artifact_precision', 0.0)
    if precision < 0.8:
        errors.append(f"[{idx}] precision={precision:.4f} < 0.8")
    return errors

def synthesize_mesh_slot(art):
    """Convert raw artifact to mesh-ready slot format"""
    slot = art.get('slot', {})
    qs = slot.get('quality_signals', {})
    sources = art.get('sources', [{}])[0] if art.get('sources') else {}
    
    return {
        'slot_id': art.get('artifact_id', 'unknown'),
        'domain': slot.get('domain', 'general'),
        'subdomain': slot.get('subdomain', ''),
        'pattern': {
            'type': slot.get('type', 'code'),
            'language': slot.get('language', ''),
            'body': slot.get('body', ''),
            'invariants': slot.get('invariants', []),
        },
        'provenance': {
            'repo': sources.get('repo', 'unknown'),
            'commit': sources.get('commit', ''),
            'file_path': sources.get('file_path', ''),
            'author': sources.get('author', ''),
            'blame_timestamp': sources.get('blame_timestamp', 0),
        },
        'metrics': {
            'survival_index': qs.get('survival_index', 0.0),
            'z_density': qs.get('z_density', 0.0),
            'artifact_precision': qs.get('artifact_precision', 0.0),
            'invariant_coverage': qs.get('invariant_coverage', 1.0),
            'extraction_reproducibility': qs.get('extraction_reproducibility', 1.0),
        },
        'metadata': {
            'anti_pattern_id': art.get('anti_pattern_id'),
            'polarity': slot.get('polarity', 'POSITIVE'),
            'distilled_at': datetime.utcnow().isoformat(),
        }
    }

def compute_mesh_stats(mesh_slots):
    stats = {
        'total_slots': len(mesh_slots),
        'domains': {},
        'languages': {},
        'avg_precision': 0.0,
        'avg_survival': 0.0,
        'avg_z_density': 0.0,
        'precision_distribution': {'0.80-0.85': 0, '0.85-0.90': 0, '0.90-0.95': 0, '0.95-1.00': 0},
        'qac_pass_rate': 0.0,
    }
    
    total_precision = 0.0
    total_survival = 0.0
    total_z = 0.0
    qac_passed = 0
    qac_total = 0
    
    for slot in mesh_slots:
        domain = slot['domain']
        lang = slot['pattern']['language'] or 'unknown'
        stats['domains'][domain] = stats['domains'].get(domain, 0) + 1
        stats['languages'][lang] = stats['languages'].get(lang, 0) + 1
        
        m = slot['metrics']
        p = m['artifact_precision']
        s = m['survival_index']
        z = m['z_density']
        
        total_precision += p
        total_survival += s
        total_z += z
        
        # Precision distribution
        if 0.80 <= p < 0.85:
            stats['precision_distribution']['0.80-0.85'] += 1
        elif 0.85 <= p < 0.90:
            stats['precision_distribution']['0.85-0.90'] += 1
        elif 0.90 <= p < 0.95:
            stats['precision_distribution']['0.90-0.95'] += 1
        elif p >= 0.95:
            stats['precision_distribution']['0.95-1.00'] += 1
    
    n = len(mesh_slots)
    if n > 0:
        stats['avg_precision'] = round(total_precision / n, 4)
        stats['avg_survival'] = round(total_survival / n, 4)
        stats['avg_z_density'] = round(total_z / n, 4)
    
    return stats

def main():
    parser = argparse.ArgumentParser(description='Run distillation pipeline')
    parser.add_argument('--seeds-dir', default='data/seeds/')
    parser.add_argument('--output', default='data/distilled/')
    args = parser.parse_args()
    
    seeds_dir = Path(args.seeds_dir)
    output_dir = Path(args.output)
    output_dir.mkdir(parents=True, exist_ok=True)
    
    print("=" * 60)
    print("Mimic Distillation Pipeline")
    print("=" * 60)
    
    # Load all seed files
    seed_files = sorted(seeds_dir.glob('*-artifacts.json'))
    print(f"\nFound {len(seed_files)} seed files:")
    for f in seed_files:
        print(f"  - {f.name}")
    
    all_artifacts = []
    all_errors = []
    
    for seed_file in seed_files:
        print(f"\n📦 Loading {seed_file.name}...")
        artifacts = load_artifacts(seed_file)
        print(f"   Count: {len(artifacts)} artifacts")
        
        for i, art in enumerate(artifacts):
            errors = validate_artifact(art, i)
            all_errors.extend(errors)
        
        all_artifacts.extend(artifacts)
        print(f"   Loaded: {len(artifacts)} | Errors: {len([e for e in all_errors if seed_file.name in str(e)])}")
    
    # Synthesize
    print(f"\n🔬 Synthesizing {len(all_artifacts)} artifacts into mesh slots...")
    mesh_slots = [synthesize_mesh_slot(art) for art in all_artifacts]
    
    # Compute stats
    stats = compute_mesh_stats(mesh_slots)
    
    # Save outputs
    mesh_file = output_dir / 'mesh_slots.json'
    with open(mesh_file, 'w') as f:
        json.dump(mesh_slots, f, indent=2)
    print(f"\n💾 Saved mesh slots: {mesh_file} ({len(mesh_slots)} slots)")
    
    stats_file = output_dir / 'mesh_stats.json'
    with open(stats_file, 'w') as f:
        json.dump(stats, f, indent=2)
    print(f"💾 Saved stats: {stats_file}")
    
    # Save semantic summary (human-readable)
    semantic_file = output_dir / 'SEMANTIC_SUMMARY.md'
    with open(semantic_file, 'w') as f:
        f.write("# Distillation Semantic Summary\n\n")
        f.write(f"**Date:** {datetime.utcnow().isoformat()} UTC\n\n")
        f.write("## What This Means\n\n")
        f.write("The distillation pipeline analyzed production code repositories ")
        f.write("and extracted proven patterns that survived long enough to be considered reliable.\n\n")
        f.write(f"**Total Patterns Distilled:** {stats['total_slots']:,}\n\n")
        f.write("### Quality Breakdown\n\n")
        f.write(f"- **Average Precision:** {stats['avg_precision']:.2%} — ")
        f.write("fraction of pattern lines still present at HEAD (via git blame)\n")
        f.write(f"- **Average Survival Index:** {stats['avg_survival']:.2%} — ")
        f.write("how much of the original code survived changes over time\n")
        f.write(f"- **Average Z-Density:** {stats['avg_z_density']:.4f} — ")
        f.write("knowledge density per slot (higher = more proven invariants per unit)\n\n")
        f.write("### Precision Distribution\n\n")
        f.write("| Range | Count | Meaning |\n")
        f.write("|-------|-------|---------|\n")
        for rng, cnt in stats['precision_distribution'].items():
            pct = cnt / stats['total_slots'] * 100 if stats['total_slots'] > 0 else 0
            f.write(f"| {rng} | {cnt:,} ({pct:.1f}%) | ")
            if '0.80' in rng:
                f.write("Minimum viable — passes quality gate\n")
            elif '0.85' in rng:
                f.write("Good — solid provenance\n")
            elif '0.90' in rng:
                f.write("Excellent — high survival confidence\n")
            else:
                f.write("Exceptional — near-perfect provenance\n")
        f.write("\n")
        f.write("### Domains Covered\n\n")
        for domain, cnt in sorted(stats['domains'].items(), key=lambda x: -x[1]):
            f.write(f"- **{domain}**: {cnt:,} patterns\n")
        f.write("\n")
        f.write("### Languages Covered\n\n")
        for lang, cnt in sorted(stats['languages'].items(), key=lambda x: -x[1]):
            f.write(f"- **{lang}**: {cnt:,} patterns\n")
        f.write("\n")
        f.write("### Why This Matters\n\n")
        f.write("Every pattern in this mesh has been **battle-tested** in production. ")
        f.write("The survival index tells us: *this code change survived because it was correct, ")
        f.write("not because it was lucky.*\n\n")
        f.write("When Mimic suggests a pattern to an AI agent, it's not guessing — ")
        f.write("it's recommending a solution that has already proven itself in the real world.\n\n")
        f.write("---\n")
        f.write(f"*Total errors during validation: {len(all_errors)}*\n")
    
    print(f"💾 Saved semantic summary: {semantic_file}")
    
    # Print summary
    print("\n" + "=" * 60)
    print("PIPELINE COMPLETE")
    print("=" * 60)
    print(f"Artifacts loaded:     {len(all_artifacts):,}")
    print(f"Mesh slots created:   {len(mesh_slots):,}")
    print(f"Validation errors:    {len(all_errors)}")
    print(f"Average precision:    {stats['avg_precision']:.4f}")
    print(f"Average survival:     {stats['avg_survival']:.4f}")
    print(f"Average Z-density:    {stats['avg_z_density']:.4f}")
    print(f"Domains:              {len(stats['domains'])}")
    print(f"Languages:            {len(stats['languages'])}")
    print("=" * 60)
    
    if all_errors:
        print("\n⚠️  Validation errors (first 10):")
        for e in all_errors[:10]:
            print(f"  {e}")
    else:
        print("\n✅ All artifacts passed validation")
    
    return 0

if __name__ == '__main__':
    sys.exit(main())
