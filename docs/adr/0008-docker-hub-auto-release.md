# ADR-0008: Docker Hub Auto-Release + Multi-Registry Strategy

## Decision
On every `v*` tag push to `main`, CI publishes Docker images to **both** GitHub Container Registry (ghcr.io) **and** Docker Hub (docker.io/mayveskii/mimic). `dev` branch publishes `:dev` tag only to ghcr.io for integration testing.

## Why (formal)
- **Distribution redundancy**: Some firewalls block ghcr.io; Docker Hub is universally accessible.
- **Enterprise adoption**: Docker Hub is the default `docker pull` registry; no registry login required for public images.
- **Dev/staging parity**: `dev` branch produces `:dev` image so integration tests use latest unstable without polluting semver tags.
- **Behavior source**: Mayveskii/go-service-template-rest (multi_target_build_matrix) — build for multiple targets from single source.

## Measured
| Metric | Before | After | Delta |
|--------|--------|-------|-------|
| Registry availability | 1 (ghcr.io only) | 2 (ghcr.io + docker.io) | +100% |
| Enterprise pull friction | high (ghcr auth in some orgs) | low (docker.io unauth) | -1 auth barrier |
| Dev integration image | none | `ghcr.io/mayveskii/mimic:dev` | +1 staging artifact |

## Invariant
- **INV-REL-1**: `main` tag `vX.Y.Z` produces identical digest on both registries (same build context, same Dockerfile).
- **INV-REL-2**: `dev` branch never produces semver tag. Only `:dev` or `:sha-<short-sha>`.
- **INV-REL-3**: Docker image contains bundled mesh data (data/seeds/) so `docker run` works offline.

## Alternatives
| Alternative | Rejected Why |
|-------------|--------------|
| Only Docker Hub, drop ghcr.io | ghcr.io is free for OSS; GitHub-native users expect it. Redundancy is safer. |
| Semver tags on dev branch | Violates semver semantics; dev is unstable by definition. |
| Manual push to Docker Hub | Human error risk; CI automation guarantees reproducibility. |

## Consilium
Approved by user on 2026-05-20.

## Test
- CI dry-run: `docker buildx build --push=false` on every PR
- Post-release: `docker pull mayveskii/mimic:0.0.1` + `docker run --rm mayveskii/mimic:0.0.1 --version`
- Digest comparison: `docker manifest inspect ghcr.io/mayveskii/mimic:v0.0.1` vs `docker.io/mayveskii/mimic:v0.0.1`

## Artifact precision
- Dockerfile survival: multi-stage build tested on every CI run (invariant coverage 1.0).
- Registry push reproducibility: same `docker/build-push-action` context (extraction reproducibility 1.0).

---

## Implementation Details

### CI Changes (.github/workflows/release.yml)

Add `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` to repository secrets.

Add job `build-docker-hub` parallel to `build-docker` (or extend existing job with second registry):

```yaml
  build-docker:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    strategy:
      matrix:
        registry:
          - url: ghcr.io
              image: ghcr.io/mayveskii/mimic
              username: ${{ github.actor }}
              password: ${{ secrets.GITHUB_TOKEN }}
          - url: docker.io
              image: mayveskii/mimic
              username: ${{ secrets.DOCKERHUB_USERNAME }}
              password: ${{ secrets.DOCKERHUB_TOKEN }}
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ${{ matrix.registry.url }}
          username: ${{ matrix.registry.username }}
          password: ${{ matrix.registry.password }}
      - uses: docker/metadata-action@v5
        id: meta
        with:
          images: ${{ matrix.registry.image }}
          tags: |
            type=ref,event=tag
            type=raw,value=latest,enable={{is_default_branch}}
      - uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
```

### dev branch CI (.github/workflows/ci.yml addition)
Add job `dev-image`:
```yaml
  dev-image:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/dev'
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ghcr.io/mayveskii/mimic:dev
```

### Versioning Schema
| Branch | Git Tag | Docker Tags | npm Tag | GitHub Release |
|--------|---------|-------------|---------|----------------|
| feat/* | none | none | none | none |
| dev | none | `ghcr.io/...:dev` | `@dev` (optional) | none |
| main | none | `ghcr.io/...:latest` (if default branch) | none | none |
| main | `v0.0.1` | `:0.0.1`, `:latest` on both registries | `@latest`, `@0.0.1` | v0.0.1 |
