#!/usr/bin/env bash
#
# distill.sh — Mimic Code Distillation Pipeline
#
# Clones a production repo, runs git blame, computes survival index,
# extracts high-survival code patterns, encodes as atomic artifacts,
# writes to bmap seed files.
#
# Usage: ./distill.sh <org>/<repo> [commit_sha]
#
# Reproducibility: same repo + same commit → same output (bitwise-identical)
#
# Phases: clone → blame → survival → filter → extract → encode → slot
#
set -euo pipefail

REPO="${1:?Usage: distill.sh <org>/<repo> [commit_sha]}"
COMMIT="${2:-HEAD}"
SI_THRESHOLD="${SI_THRESHOLD:-0.7}"
Z_THRESHOLD="${Z_THRESHOLD:-0.3}"
MAX_TOKENS="${MAX_TOKENS:-2048}"
SEEDS_DIR="${SEEDS_DIR:-data/seeds}"
CACHE_DIR="${CACHE_DIR:-/tmp/mimic-distill}"
TOOL_VERSION="distill.sh-1.0.0"

REPO_SAFE="${REPO//\//-}"
WORK_DIR="${CACHE_DIR}/${REPO_SAFE}"
SEED_FILE="${SEEDS_DIR}/${REPO_SAFE}.bmap"

echo "[distill] repo=${REPO} commit=${COMMIT} si_threshold=${SI_THRESHOLD}"

mkdir -p "${WORK_DIR}" "${SEEDS_DIR}"

# --- Phase 1: Clone ---
if [ ! -d "${WORK_DIR}/.git" ]; then
    echo "[distill] cloning ${REPO}..."
    git clone --quiet "https://github.com/${REPO}.git" "${WORK_DIR}"
fi

cd "${WORK_DIR}"
git checkout --quiet "${COMMIT}"
ACTUAL_COMMIT="$(git rev-parse HEAD)"
echo "[distill] checked out ${ACTUAL_COMMIT}"

# --- Phase 2: Blame — compute survival index per commit ---
echo "[distill] running git blame..."
BLAME_FILE="${CACHE_DIR}/${REPO_SAFE}.blame"
> "${BLAME_FILE}"

find . -name '*.c' -o -name '*.go' -o -name '*.py' -o -name '*.rs' -o -name '*.zig' \
    -o -name '*.h' -o -name '*.hpp' | while read -r f; do
    git blame -l -t "${f}" 2>/dev/null | while read -r line; do
        echo "${f}:${line}"
    done
done >> "${BLAME_FILE}"

# --- Phase 3: Survival Index computation ---
echo "[distill] computing survival index..."
SI_FILE="${CACHE_DIR}/${REPO_SAFE}.si"
python3 "$(dirname "$0")/compute_survival.py" \
    --blame "${BLAME_FILE}" \
    --output "${SI_FILE}" \
    --threshold "${SI_THRESHOLD}"

# --- Phase 4: Filter — keep only commits with SI ≥ threshold ---
FILTERED_FILE="${CACHE_DIR}/${REPO_SAFE}.filtered"
awk -v th="${SI_THRESHOLD}" '$3 >= th {print}' "${SI_FILE}" > "${FILTERED_FILE}"
FILTERED_COUNT="$(wc -l < "${FILTERED_FILE}")"
echo "[distill] ${FILTERED_COUNT} commit-lines above SI=${SI_THRESHOLD}"

# --- Phase 5: Extract — pull out functions from surviving code ---
echo "[distill] extracting patterns..."
PATTERNS_FILE="${CACHE_DIR}/${REPO_SAFE}.patterns"
python3 "$(dirname "$0")/extract_patterns.py" \
    --filtered "${FILTERED_FILE}" \
    --repo-dir "${WORK_DIR}" \
    --output "${PATTERNS_FILE}" \
    --max-tokens "${MAX_TOKENS}"

PATTERNS_COUNT="$(wc -l < "${PATTERNS_FILE}")"
echo "[distill] ${PATTERNS_COUNT} patterns extracted"

# --- Phase 6: Encode — produce atomic artifacts ---
echo "[distill] encoding artifacts..."
python3 "$(dirname "$0")/encode_artifacts.py" \
    --patterns "${PATTERNS_FILE}" \
    --repo "${REPO}" \
    --commit "${ACTUAL_COMMIT}" \
    --tool-version "${TOOL_VERSION}" \
    --output "${SEED_FILE}" \
    --format protobuf

# --- Phase 7: Compute Z-density ---
echo "[distill] computing Z-density..."
Z_DENSITY="$(python3 "$(dirname "$0")/compute_zdensity.py" \
    --seed "${SEED_FILE}" \
    --repo "${REPO}" 2>/dev/null || echo '0.0')"
echo "[distill] Z-density=${Z_DENSITY}"

# --- Phase 8: Update manifest ---
echo "[distill] updating repos-manifest.yaml..."
python3 "$(dirname "$0")/update_manifest.py" \
    --manifest "$(dirname "$0")/../../mimicrya/repos-manifest.yaml" \
    --repo "${REPO}" \
    --commit "${ACTUAL_COMMIT}" \
    --slots "${PATTERNS_COUNT}" \
    --z-density "${Z_DENSITY}" \
    --status "distilled"

echo "[distill] done: ${SEED_FILE} (${PATTERNS_COUNT} artifacts, Z=${Z_DENSITY})"
