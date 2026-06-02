# Releasing Mimic

## Branches

- `main` — stable. PRs only.
- `dev` — integration. Feature branches merge here.

## Pull Requests

1. Green CI required (`lint`, `build`, `test`)
2. PR to `main` requires review
3. Do not merge if any job is red

## Versioning

- Semantic versioning: `vMAJOR.MINOR.PATCH`
- Tag created manually after merge to `main`
- Tag is immutable. Mistake → new tag

## Release Process

1. Ensure `main` is green
2. Create annotated tag:
   ```bash
   git tag -a v0.x.x -m "Release v0.x.x"
   git push origin v0.x.x
   ```
3. GoReleaser automatically:
   - Builds `linux/amd64` binary
   - Creates GitHub Release
   - Publishes Docker image to GHCR

## Data Pipeline

- Weekly (or manual) — workflow `data.yml`
- Creates branch `auto/data-sync-YYYYMMDD`
- Opens PR to `dev`
- You decide: merge or close
- No auto-commits to `main`

## npm

- Not active in current releases
- Will be enabled when Mimic is ready for wide distribution
- See `.goreleaser.yml` footer for future usage
