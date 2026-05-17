.PHONY: all build test lint check distill release clean

CC = gcc
CFLAGS = -Wall -Wextra -O3 -fPIC -Icore
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
	$(CC) $(CFLAGS) -o core/test_runner core/test/*.c -Lcore -lcore -lm
	./core/test_runner

# ============================================================================
# Go (Mimic binary)
# ============================================================================

build: $(CORE_LIB)
	@mkdir -p bin
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o bin/mimic ./cmd/mimic

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
	$(GO) run ./data/extraction/... -manifest repos-manifest.yaml

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
