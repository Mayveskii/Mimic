# Runtime image for Mimic MCP Server
# Binary is built by GoReleaser and copied into this image

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libssl3 \
    git \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m -s /bin/bash -u 1000 mimic

COPY mimic /usr/local/bin/mimic

LABEL org.opencontainers.image.title="Mimic MCP Server" \
      org.opencontainers.image.description="Deterministic AI-agent tool orchestration with C-core execution engine" \
      org.opencontainers.image.source="https://github.com/Mayveskii/Mimic" \
      org.opencontainers.image.licenses="MIT"

ENV MIMIC_PORT=1337

EXPOSE 1337

USER mimic

ENTRYPOINT ["mimic"]
CMD ["serve"]
