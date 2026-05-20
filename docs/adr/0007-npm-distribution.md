# ADR-0007: npm Package Distribution (npx mimic)

## Decision
Distribute Mimic binary via npm package `@mayveskii/mimic` with platform-detecting install script, enabling `npx @mayveskii/mimic serve` as zero-config entrypoint.

## Why (formal)
- **Distribution reach**: npm is the default package manager for AI-agent ecosystems (Cursor, Claude Code, Copilot extensions).
- **Zero install friction**: `npx` downloads and caches automatically; no `curl | bash` required for JS/TS agents.
- **Version parity**: npm version = Git tag = Docker tag. Single source of truth.
- **Behavior source**: Mayveskii/go-service-template-rest (bootstrap lifecycle) — phased startup with version detection.

## Measured
| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| Install steps (new user) | 3 (clone, make, configure) | 1 (`npx`) | -67% |
| Platform coverage | manual | auto-detect (linux/darwin/win, amd64/arm64) | +3 platforms |
| Agent adoption barrier | high (needs shell) | low (needs only node) | -1 dep class |

## Invariant
- **INV-DIST-1**: npm package MUST NOT bundle binary. It downloads platform-matched binary from GitHub Release on first run (survival index of install logic ≥ 0.9).
- **INV-DIST-2**: If GitHub Release unavailable → graceful fallback to build-from-source with clear error.
- **INV-DIST-3**: npm version string == Git tag == Docker image tag (semantic consistency).

## Alternatives
| Alternative | Rejected Why |
|-------------|--------------|
| Bundle binary in npm (all platforms) | Package size 45MB+; violates npm ecosystem norms. |
| Publish only to GitHub Releases | Excludes AI tools that prefer `npx` (Copilot, Cursor). |
| Homebrew formula only | macOS-only; no Windows/Linux parity. |

## Consilium
Approved by user on 2026-05-20 as part of distribution strategy.

## Test
- CI job: `npm publish --dry-run` on every tag push
- Manual: `npx @mayveskii/mimic@dev --version` must print correct semver
- Platform matrix: ubuntu-latest, macos-latest, windows-latest

## Artifact precision
- Install script survival: downloaded binary must match SHA256 from release asset (hash verification = invariant coverage 1.0).
- Extraction reproducibility: 100% (same GitHub Release asset for all platforms).

---

## Implementation Details

### package.json schema
```json
{
  "name": "@mayveskii/mimic",
  "version": "0.0.1",
  "description": "Mimic MCP Server — deterministic AI-agent tool orchestration",
  "bin": {
    "mimic": "./bin/mimic.js"
  },
  "scripts": {
    "postinstall": "node scripts/install.js",
    "prepublishOnly": "npm test"
  },
  "files": [
    "bin/",
    "scripts/",
    "README.md",
    "LICENSE"
  ],
  "engines": { "node": ">=18.0.0" },
  "os": [ "darwin", "linux", "win32" ],
  "cpu": [ "x64", "arm64" ],
  "publishConfig": { "access": "public" }
}
```

### install.js logic (postinstall)
1. Detect `process.platform` + `process.arch`
2. Map to release asset: `mimic-{os}-{arch}`
3. Check `MIMIC_VERSION` env → default to `package.json version`
4. Download from `https://github.com/Mayveskii/Mimic/releases/download/v${version}/${asset}`
5. Verify SHA256 against `.sha256` asset
6. Write to `bin/mimic-native` (or `bin/mimic-native.exe` on Win)
7. `chmod +x` on Unix

### bin/mimic.js wrapper
- Spawns `bin/mimic-native` with all `process.argv.slice(2)` forwarded
- If native binary missing → triggers `install.js` inline → then spawns

### CI Integration (.github/workflows/release.yml)
Add job `publish-npm` after `build-binaries`:
- needs: `[build-binaries, test]`
- steps: checkout → setup-node → npm publish (needs `NPM_TOKEN` secret)
- Only on tags starting with `v`
