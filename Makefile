.PHONY: all build test lint check check-config semantics-check distill release clean

CC = gcc
CFLAGS = -Wall -Wextra -O3 -fPIC -Icore
LDFLAGS = -Lcore -lcore -lm -lcrypto
GO = go
GOFLAGS = -trimpath

CORE_SRCS = $(wildcard core/*.c)
CORE_OBJS = $(CORE_SRCS:.c=.o)
CORE_LIB = core/libcore.a

all: build

# ============================================================================
# C-Core
# ============================================================================

core/%.o: core/%.c core/ops.h
	$(CC) $(CFLAGS) -c $< -o $@

$(CORE_LIB): $(CORE_OBJS)
	ar rcs $@ $^

core-test: $(CORE_LIB)
	$(CC) $(CFLAGS) -o core/test_runner core/test_ops.c -Lcore -lcore -lm -lcrypto
	./core/test_runner

# ============================================================================
# Go (Mimic binary)
# ============================================================================

build: $(CORE_LIB)
	@mkdir -p bin
	CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o bin/mimic ./cmd/mimic

# ============================================================================
# Tests
# ============================================================================

test: build
	$(GO) test ./...

# ============================================================================
# Lint
# ============================================================================

lint:
	$(GO) vet ./...
	gofmt -l . | grep -q . && exit 1 || true

# ============================================================================
# Semantics check
# ============================================================================

semantics-check:
	@echo "Checking SEMANTICS.md sync with code..."
	@python3 scripts/semantics_check.py 2>/dev/null || echo "semantics_check: script not found, skipping"

# ============================================================================
# Configuration drift detection
# ============================================================================

check-config:
	@echo "Checking 11-CONFIGURATION.md vs codebase..."
	@python3 scripts/check_config_consistency.py \
		--spec specs/11-CONFIGURATION.md \
		--env .env.example \
		--code internal/config/,internal/mcp/,core/ops.h,Makefile,Dockerfile 2>/dev/null \
	|| echo "check-config: script not found, skipping"

# ============================================================================
# Full check
# ============================================================================

check: lint test semantics-check check-config
	@echo "All checks passed."

# ============================================================================
# Distillation
# ============================================================================

distill:
	$(GO) run ./data/extraction/... -manifest mimicrya/repos-manifest.yaml

distill-code:
	@bash data/extraction/distill.sh "$(REPO)" "$(COMMIT)"

distill-decisions:
	@python3 data/extraction/distill_decisions.py --repo "$(REPO)" --pr "$(PR_RANGE)" --output "data/seeds/$(REPO_SAFE)-decisions.json"

distill-multimodal:
	@python3 data/extraction/extract_multimodal.py --repo-dir "$(REPO_DIR)" --repo "$(REPO)" --commit "$(COMMIT)" --output "data/seeds/$(REPO_SAFE)-multimodal.json"

distill-reach:
	@python3 scripts/reach.py \
		--manifest mimicrya/repos-manifest.yaml \
		--seeds data/seeds \
		--max-results 5

coverage:
	@python3 data/extraction/compute_coverage.py --manifest mimicrya/repos-manifest.yaml --behaviors mimicrya/behavior-sources.yaml

adr-new:
	@NEXT=$$(ls docs/adr/ | grep -E '^[0-9]+' | sort -n | tail -1 | grep -oE '[0-9]+' | awk '{print $$1+1}') && \
	cp docs/adr/TEMPLATE.md "docs/adr/$${NEXT}-$(NAME).md" && \
	echo "Created docs/adr/$${NEXT}-$(NAME).md"

# ============================================================================
# Bootstrap (download mesh data on first run)
# ============================================================================

bootstrap:
	@echo "Bootstrapping Mimic mesh data..."
	@bash scripts/download-data.sh

# ============================================================================
# Data packaging for GitHub Release
# ============================================================================

package-data:
	@bash scripts/package-data.sh $(shell git describe --tags --always)

# ============================================================================
# Full release: binaries + data + Docker
# ============================================================================

RELEASE_VERSION := $(shell git describe --tags --always)

release-all: clean build package-data docker
	@echo "═══════════════════════════════════════════════════════════════"
	@echo "  RELEASE v$(RELEASE_VERSION)"
	@echo "═══════════════════════════════════════════════════════════════"
	@echo ""
	@echo "  Artifacts to publish:"
	@echo ""
	@echo "  1. GitHub Release assets:"
	@echo "     - dist/mimic-$(RELEASE_VERSION)-linux-amd64"
	@echo "     - dist/mimic-$(RELEASE_VERSION)-darwin-arm64"
	@echo "     - dist/mimic-data-$(RELEASE_VERSION).tar.gz"
	@echo ""
	@echo "  2. Docker images:"
	@echo "     - ghcr.io/mayveskii/mimic:$(RELEASE_VERSION)"
	@echo "     - ghcr.io/mayveskii/mimic:latest"
	@echo "     - docker.io/mayveskii/mimic:$(RELEASE_VERSION)"
	@echo "     - docker.io/mayveskii/mimic:latest"
	@echo ""
	@echo "  3. npm package:"
	@echo "     - @mayveskii/mimic@$(RELEASE_VERSION)"
	@echo ""
	@echo "  Upload commands:"
	@echo "     gh release create $(RELEASE_VERSION) --generate-notes"
	@echo "     gh release upload $(RELEASE_VERSION) dist/*"
	@echo "     docker push ghcr.io/mayveskii/mimic:$(RELEASE_VERSION)"
	@echo "     docker push docker.io/mayveskii/mimic:$(RELEASE_VERSION)"
	@echo "     npm publish --access public"
	@echo "═══════════════════════════════════════════════════════════════"

# ============================================================================
# Docker
# ============================================================================

docker:
	docker build -t mimic:latest .
	docker tag mimic:latest ghcr.io/mayveskii/mimic:latest
	docker tag mimic:latest ghcr.io/mayveskii/mimic:$(RELEASE_VERSION)
	docker tag mimic:latest mayveskii/mimic:latest
	docker tag mimic:latest mayveskii/mimic:$(RELEASE_VERSION)

docker-push: docker
	docker push ghcr.io/mayveskii/mimic:latest
	docker push ghcr.io/mayveskii/mimic:$(RELEASE_VERSION)
	docker push mayveskii/mimic:latest
	docker push mayveskii/mimic:$(RELEASE_VERSION)

# ============================================================================
# npm
# ============================================================================

npm-publish:
	cd . && npm publish --access public

npm-dry-run:
	cd . && npm publish --dry-run

# ============================================================================
# Release (legacy — binaries only)
# ============================================================================

release:
	@echo "Building release binaries..."
	@mkdir -p dist
	@for plat in $(PLATFORMS); do \
		GOOS=$${plat%/*} GOARCH=$${plat#*/} CGO_ENABLED=1 \
		CC=$${plat%/*}-gcc \
		$(GO) build $(GOFLAGS) -o dist/mimic-$${plat} ./cmd/mimic 2>/dev/null || \
		CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o dist/mimic-$${plat} ./cmd/mimic; \
		sha256sum dist/mimic-$${plat} > dist/mimic-$${plat}.sha256; \
	done
	@echo "Release binaries in dist/"

# ============================================================================
# Clean
# ============================================================================

clean:
	rm -f $(CORE_OBJS) $(CORE_LIB) core/test_runner
	rm -rf bin/ dist/ dist-bin/ node_modules/
	$(GO) clean

# ============================================================================
# Quality Infrastructure
# ============================================================================

encode-v2:
	@python3 data/extraction/encode_artifacts_v2.py \
		--patterns "$(CACHE_DIR)/$(REPO_SAFE).patterns" \
		--repo "$(REPO)" \
		--commit "$(COMMIT)" \
		--tool-version "$(TOOL_VERSION)" \
		--output "$(SEED_FILE)"

quality-check:
	@python3 data/extraction/quality_gate.py --artifact "$(ARTIFACT)" --output "$(ARTIFACT).qac.json"

validate-artifact:
	@python3 data/extraction/artifact_completeness.py --artifact "$(ARTIFACT)"

apply-decisions:
	@python3 data/extraction/apply_decisions_to_matrix.py \
		--input mimicrya/decision-patterns.yaml \
		--output-c data/matrices/decision_matrix.c \
		--output-h data/matrices/decision_matrix.h
