# Multi-stage build for Mimic MCP Server
# Stage 1: Build C-core + Go binary with CGO
# Stage 2: Minimal runtime with SSL support
#
# Default ports (1337-style):
#   1337 - Main MCP stdio/SSE server (default, most frequent)
#   1117 - HTTP API / metrics endpoint
#   1227 - Secondary / admin endpoint
#   1447 - WebSocket endpoint
#   1557 - Internal mesh communication
#
# All ports can be overridden via ENV variables at runtime

## Stage 1: Builder
FROM golang:1.22-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    make \
    libssl-dev \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy dependency files first (for layer caching)
COPY go.mod ./
COPY Makefile ./

# Copy source code
COPY core/ core/
COPY internal/ internal/
COPY cmd/ cmd/
COPY data/ data/
COPY mimicrya/ mimicrya/
COPY specs-v2/ specs-v2/
COPY specs/ specs/
COPY .github/ .github/

# Build: C-core library + Go binary
# CGO must be enabled for C-core bridge
RUN make clean && CGO_ENABLED=1 make

## Stage 2: Runtime
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libssl3 \
    git \
    make \
    gcc \
    golang-go \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN useradd -m -s /bin/bash -u 1000 mimic

# Copy binary from builder
COPY --from=builder /build/bin/mimic /usr/local/bin/mimic

# Copy specs and documentation (for reference)
COPY --from=builder /build/specs-v2/ /usr/share/mimic/specs/
COPY --from=builder /build/README.md /usr/share/mimic/
COPY --from=builder /build/AGENTS.md /usr/share/mimic/

# Health check on main port
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD mimic ping || exit 1

# Metadata
LABEL org.opencontainers.image.title="Mimic MCP Server" \
      org.opencontainers.image.description="Deterministic AI-agent tool orchestration with C-core execution engine" \
      org.opencontainers.image.source="https://github.com/Mayveskii/Mimic" \
      org.opencontainers.image.version="0.1.0" \
      org.opencontainers.image.licenses="MIT"

# Environment variables for port configuration
# All ports use 1337-style numbering (leet speak aesthetic)
ENV MIMIC_PORT=1337 \
    MIMIC_HTTP_PORT=1117 \
    MIMIC_ADMIN_PORT=1227 \
    MIMIC_WS_PORT=1447 \
    MIMIC_MESH_PORT=1557 \
    MIMIC_LOG_LEVEL=info \
    MIMIC_BUDGET_TOKENS=100000 \
    MIMIC_BUDGET_TIME_MS=3600000 \
    MIMIC_MAX_TASKS=50

# Expose all configured ports
# 1337 - Main MCP server (stdio/SSE/HTTP fallback)
# 1117 - HTTP REST API and Prometheus metrics
# 1227 - Admin/management API
# 1447 - WebSocket transport
# 1557 - Internal mesh node communication (future)
EXPOSE 1337 1117 1227 1447 1557

USER mimic

ENTRYPOINT ["mimic"]
CMD ["serve"]
