# Deployment Guide — Mimic v0.3

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Client    │────▶│  Mimic TCP   │────▶│   C-Core     │
│  (opencode) │◄────│   :1337      │◄────│  (libcore.a) │
└─────────────┘     └──────────────┘     └──────────────┘
        │                   │
        │            ┌──────┴──────┐
        │            │  Mesh       │
        │            │  149K slots │
        │            └──────┬──────┘
        │                   │
   ┌────┴─────┐      ┌─────┴──────┐
   │ Embed    │      │  Qdrant    │
   │ :1137    │      │  :6333     │
   │(systemd) │      │  (docker)  │
   └──────────┘      └────────────┘
```

## Production Stack

### 1. Embed Service (systemd)

```ini
# /etc/systemd/system/mimic-embed.service
[Unit]
Description=Mimic TextEmbedding Service
After=network.target

[Service]
Type=simple
ExecStart=/opt/mimic/bin/embed-server
WorkingDirectory=/opt/mimic
Environment=MODEL=all-MiniLM-L6-v2
Environment=PORT=1137
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable --now mimic-embed
# Verify: curl http://localhost:1137/health
```

### 2. Qdrant (Docker)

```bash
docker run -d --name qdrant \
  -p 6333:6333 \
  -v /opt/mimic/data/qdrant:/qdrant/storage \
  qdrant/qdrant:latest

# Verify: curl http://localhost:6333/healthz
```

Collection: `binary_mesh_chunks`
- 180,684 points (384-dim, Cosine)
- HNSW index (ef=128, m=16)

### 3. Mimic Server (systemd)

```ini
# /etc/systemd/system/mimic.service
[Unit]
Description=Mimic MCP Server (v0.3)
After=network.target mimic-embed.service
Wants=mimic-embed.service

[Service]
Type=simple
ExecStart=/opt/mimic/bin/mimic serve --tcp :1337
WorkingDirectory=/opt/mimic
Environment=MIMIC_WORKING_DIR=/opt/mimic
Environment=MIMIC_MESH_DIR=/opt/mimic/data/mesh/graphs
Environment=MIMIC_EMBED_ENDPOINT=http://localhost:1137
Environment=MIMIC_TCP_ADDR=:1337
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**Critical:** TCP mode pre-loads mesh ONCE (1.9GB RSS). `WithTransport()` clones per connection — no duplicate loads.

### 4. Bootstrap Script

```bash
#!/bin/bash
# /opt/mimic/scripts/bootstrap.sh
set -e

MIMIC_DIR="/opt/mimic"
BIN_DIR="$MIMIC_DIR/bin"
DATA_DIR="$MIMIC_DIR/data"

# Step 1: Embed service
curl -sf http://localhost:1137/health || {
    echo "Starting embed service..."
    systemctl start mimic-embed
    sleep 5
}

# Step 2: Qdrant
curl -sf http://localhost:6333/healthz || {
    echo "Starting qdrant..."
    docker start qdrant || docker run -d --name qdrant -p 6333:6333 \
        -v "$DATA_DIR/qdrant:/qdrant/storage" qdrant/qdrant:latest
    sleep 5
}

# Step 3: Verify mesh data
if [ ! -d "$DATA_DIR/mesh/graphs" ]; then
    echo "ERROR: Mesh data not found at $DATA_DIR/mesh/graphs"
    exit 1
fi

# Step 4: Start Mimic
systemctl restart mimic
sleep 2

# Verify
TOOLS_COUNT=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
    nc -q 1 localhost 1337 | jq '.result.tools | length')
echo "Mimic ready: $TOOLS_COUNT tools available"
```

## Resource Requirements

| Component | RAM | CPU | Disk | Notes |
|-----------|-----|-----|------|-------|
| Embed | 2GB | 2 cores | 500MB | PyTorch + model |
| Qdrant | 512MB | 1 core | 5GB | HNSW indices |
| Mimic (TCP) | 2GB | 1 core | 100MB | Mesh pre-loaded |
| **Total** | **~5GB** | **4 cores** | **6GB** | For 180K vectors |

## Docker Compose (all-in-one)

```yaml
version: "3.8"
services:
  qdrant:
    image: qdrant/qdrant:latest
    ports: ["6333:6333"]
    volumes:
      - ./data/qdrant:/qdrant/storage

  embed:
    image: mimic-embed:latest
    ports: ["1137:1137"]
    environment:
      - MODEL=all-MiniLM-L6-v2
      - PORT=1137

  mimic:
    image: mimic:latest
    ports: ["1337:1337"]
    environment:
      - MIMIC_EMBED_ENDPOINT=http://embed:1137
      - MIMIC_TCP_ADDR=:1337
    depends_on:
      - qdrant
      - embed
    volumes:
      - ./data/mesh:/opt/mimic/data/mesh:ro
```

## Monitoring

```bash
# Check all services
systemctl is-active mimic mimic-embed
docker ps --filter name=qdrant

# Metrics
curl -s http://localhost:1337/metrics  # future: Prometheus
journalctl -u mimic -f
```
