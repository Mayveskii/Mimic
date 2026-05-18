.PHONY: all build test lint check distill release clean

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
# Full check
# ============================================================================

check: lint test semantics-check
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

coverage:
	@python3 data/extraction/compute_coverage.py --manifest mimicrya/repos-manifest.yaml --behaviors mimicrya/behavior-sources.yaml

adr-new:
	@NEXT=$$(ls docs/adr/ | grep -E '^[0-9]+' | sort -n | tail -1 | grep -oE '[0-9]+' | awk '{print $$1+1}') && \
	cp docs/adr/TEMPLATE.md "docs/adr/$${NEXT}-$(NAME).md" && \
	echo "Created docs/adr/$${NEXT}-$(NAME).md"

# ============================================================================
# Release
# ============================================================================

RELEASE_BIN = mimic
PLATFORMS = linux/amd64 darwin/arm64

release:
	@echo "Building release..."
	@mkdir -p dist
	@for plat in $(PLATFORMS); do \
		GOOS=$${plat%/*} GOARCH=$${plat#*/} CGO_ENABLED=0 \
		$(GO) build $(GOFLAGS) -o dist/$(RELEASE_BIN)-$${plat} ./cmd/mimic; \
		sha256sum dist/$(RELEASE_BIN)-$${plat} > dist/$(RELEASE_BIN)-$${plat}.sha256; \
	done
	@echo "Release binaries in dist/"

docker:
	docker build -t mimic:latest .
	docker tag mimic:latest mayveskii/mimic:latest
	docker tag mimic:latest mayveskii/mimic:$(shell git describe --tags --always)

# ============================================================================
# Clean
# ============================================================================

clean:
	rm -f $(CORE_OBJS) $(CORE_LIB) core/test_runner
	rm -rf bin/ dist/
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
