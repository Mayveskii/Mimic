#!/bin/bash
set -euo pipefail

echo "=== Exa Local Test ==="
echo ""

# Load secrets from project_context_main
cd /home/cisco/mimic
if [ -f project_context_main/secrets/tokens.env ]; then
  # Export only EXA_API_KEY
  export EXA_API_KEY=$(grep '^EXA_API_KEY=' project_context_main/secrets/tokens.env | cut -d= -f2)
  echo "Loaded EXA_API_KEY (len=${#EXA_API_KEY})"
else
  echo "No secrets file found. Looking for EXA_API_KEY in env..."
fi

if [ -z "${EXA_API_KEY:-}" ]; then
  echo "ERROR: EXA_API_KEY not set"
  exit 1
fi

echo ""
echo "=== 1. Direct Exa API Test (curl) ==="
curl -s -X POST https://api.exa.ai/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $EXA_API_KEY" \
  -d '{
    "query": "top GitHub repositories Go RAFT consensus 2026 stars:>100",
    "numResults": 3,
    "type": "auto",
    "includeText": ["text"]
  }' | python3 -m json.tool | head -60

echo ""
echo "=== 2. Go Build Test ==="
cd /home/cisco/mimic
go build ./...
echo "Go build: OK"

echo ""
echo "=== 3. Exa Client (Go) ==="
# Build a tiny test that imports our client (must be inside module for internal/ import)
cd /home/cisco/mimic
cat > exa_tmp_main.go << 'GOEOF'
package main

import (
	"fmt"
	"github.com/Mayveskii/Mimic/internal/tool/exa"
)

func main() {
	cfg := exa.LoadConfigFromEnv()
	fmt.Printf("Config: BaseURL=%s MaxResults=%d TimeoutMs=%d RetryMax=%d\n",
		cfg.BaseURL, cfg.MaxResults, cfg.TimeoutMs, cfg.RetryMax)
	if cfg.Disabled() {
		fmt.Println("Client DISABLED (no EXA_API_KEY)")
		return
	}
	fmt.Println("Client ENABLED")
}
GOEOF

EXA_API_KEY="$EXA_API_KEY" go run exa_tmp_main.go
rm -f exa_tmp_main.go

echo ""
echo "=== All local tests passed ==="
