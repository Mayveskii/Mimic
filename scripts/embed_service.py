#!/usr/bin/env python3
"""
Mimic Embedding Service
Loads sentence-transformers/all-MiniLM-L6-v2 and serves embeddings via HTTP.
Endpoint: POST /embed → {"text": "race condition in go"} → {"embedding": [...]}
Also provides /quantize for int8 conversion.
"""

import os
import sys
import json
import time
import math
import struct
from typing import List

try:
    from sentence_transformers import SentenceTransformer
except ImportError:
    print("ERROR: sentence-transformers not installed. Run: pip install sentence-transformers", file=sys.stderr)
    sys.exit(1)

try:
    from flask import Flask, request, jsonify
except ImportError:
    print("ERROR: flask not installed. Run: pip install flask", file=sys.stderr)
    sys.exit(1)

EMBED_DIM = 384
MODEL_NAME = os.environ.get("MIMIC_EMBED_MODEL", "sentence-transformers/all-MiniLM-L6-v2")
PORT = int(os.environ.get("MIMIC_EMBED_PORT", "1137"))

def quantize_float32_to_int8(vec: List[float]) -> List[int]:
    """Embryo-compatible quantization: scale per vector to int8 range."""
    if not vec:
        return []
    max_val = max(abs(x) for x in vec)
    if max_val < 1e-8:
        return [0] * len(vec)
    scale = 127.0 / max_val
    return [max(-128, min(127, int(round(x * scale)))) for x in vec]

# Load model
print(f"[embed] Loading model {MODEL_NAME}...", file=sys.stderr)
start = time.time()
model = SentenceTransformer(MODEL_NAME)
print(f"[embed] Model loaded in {time.time()-start:.1f}s", file=sys.stderr)

app = Flask("mimic-embed")

@app.route("/health", methods=["GET"])
def health():
    return jsonify({"status": "ok", "model": MODEL_NAME, "dim": EMBED_DIM})

@app.route("/embed", methods=["POST"])
def embed():
    req = request.get_json(force=True)
    text = req.get("text", "")
    if not text:
        return jsonify({"error": "text required"}), 400
    
    start = time.time()
    vec = model.encode(text, convert_to_numpy=True).tolist()
    elapsed = time.time() - start
    
    return jsonify({
        "embedding": vec,
        "dim": len(vec),
        "latency_ms": round(elapsed * 1000, 2)
    })

@app.route("/quantize", methods=["POST"])
def quantize():
    """Convert float32[384] to int8[384]."""
    req = request.get_json(force=True)
    vec = req.get("embedding", [])
    if len(vec) != EMBED_DIM:
        return jsonify({"error": f"expected {EMBED_DIM} dims, got {len(vec)}"}), 400
    
    int8_vec = quantize_float32_to_int8(vec)
    return jsonify({
        "int8": int8_vec,
        "hex": "".join(f"{b & 0xff:02x}" for b in struct.pack(f"<{len(int8_vec)}b", *int8_vec))
    })

@app.route("/embed_int8", methods=["POST"])
def embed_int8():
    """One-shot: text → int8[384]."""
    req = request.get_json(force=True)
    text = req.get("text", "")
    if not text:
        return jsonify({"error": "text required"}), 400
    
    start = time.time()
    vec = model.encode(text, convert_to_numpy=True).tolist()
    int8_vec = quantize_float32_to_int8(vec)
    elapsed = time.time() - start
    
    return jsonify({
        "int8": int8_vec,
        "dim": EMBED_DIM,
        "latency_ms": round(elapsed * 1000, 2)
    })

if __name__ == "__main__":
    print(f"[embed] Starting on port {PORT}", file=sys.stderr)
    app.run(host="0.0.0.0", port=PORT, threaded=True)
