```yaml
repo: Mayveskii/openmythos
url: https://github.com/Mayveskii/openmythos
language: Python
status: partial
last_sync: "2025-05-17"

description: |
  Fork of openmythos. Deep learning architecture combining Mixture-of-Experts routing with
  recurrent depth execution, depth-wise LoRA adaptation, and adaptive computation time.
  Implements prelude-recurrent-coda pattern with MoE routing via moda.py and convergence
  stability checks during training.

advantages:
  - id: om_recurrent_depth_execution
    what: Recurrent depth: same transformer block applied N times with shared weights; iterative refinement converges across depth iterations
    evidence: "main.py — recurrent depth loop applying block N times; depth_iterations config parameter controls recurrence count"

  - id: om_moe_routing
    what: Mixture-of-Experts: top-K expert selection per token with load balancing; moda.py implements expert routing + capacity factor enforcement
    evidence: "main.py — MoE layer integration; moda.py — expert selection, top-K routing, capacity_factor, auxiliary load balance loss"

  - id: om_depth_lora_adaptation
    what: Depth-wise LoRA: each recurrence iteration has its own LoRA adapter; shared base weights + per-depth low-rank deltas
    evidence: "main.py — per-depth LoRA adapter injection; lora_rank per depth level config"

  - id: om_prelude_recurrent_coda
    what: Three-phase forward pass: prelude (initial embedding) → recurrent (iterative depth blocks) → coda (final projection); structured like musical form
    evidence: "main.py — prelude module → recurrent loop → coda module; forward() follows this three-phase structure"

  - id: om_convergence_stability_check
    what: Convergence check between recurrent iterations: compute delta between iterations → if below threshold → early exit; prevents wasted compute on converged states
    evidence: "main.py — injection.get_A() computes iteration delta; convergence_threshold config; early exit when delta < threshold"

  - advantage_id: om_adaptive_computation_time
    what: ACT (Adaptive Computation Time): dynamic halting probability per token; ponder cost penalty; tokens that need more compute get more iterations
    evidence: "main.py — act_threshold config; halting probability computation per token; ponder cost added to loss"

applications:
  - advantage_id: om_recurrent_depth_execution
    implemented_in: internal/orchestrator/recurrent.go
    mechanism: "Apply same transform block N times: output_i = block(output_{i-1}); shared weights across iterations; configurable depth_iterations"
    invariant: "depth_iterations ≥ 1. Each iteration uses same block weights. Output of iteration i is input to iteration i+1."
    status: planned

  - advantage_id: om_moe_routing
    implemented_in: internal/orchestrator/moe.go
    mechanism: "Per-token router → top-K expert selection → weighted sum of expert outputs → auxiliary load balance loss"
    invariant: "Every token routed to exactly K experts. Load balance loss ≥ 0. Capacity factor limits tokens per expert."
    status: future

  - advantage_id: om_depth_lora_adaptation
    implemented_in: internal/orchestrator/lora.go
    mechanism: "Base weight W + per-depth LoRA: W + A_depth × B_depth; A_depth and B_depth are low-rank (rank r) matrices unique per depth iteration"
    invariant: "Base weights shared across all depths. Per-depth LoRA rank ≤ base rank. LoRA delta always added, never replaces."
    status: future

  - advantage_id: om_prelude_recurrent_coda
    implemented_in: internal/orchestrator/threephase.go
    mechanism: "Forward pass: prelude(embed_input) → for i in 0..N: recurrent(state) → coda(final_state) → output"
    invariant: "Prelude runs exactly once. Recurrent runs depth_iterations times. Coda runs exactly once. Order enforced."
    status: planned

  - advantage_id: om_convergence_stability_check
    implemented_in: internal/quality/convergence.go
    mechanism: "Between iterations: delta = ||output_i - output_{i-1}|| → if delta < convergence_threshold → early exit; log iteration count"
    invariant: "Convergence check after each recurrent iteration except first. Early exit preserves output_i as final state. Delta computed in same space."
    status: planned

  - advantage_id: om_adaptive_computation_time
    implemented_in: internal/orchestrator/act.go
    mechanism: "Per-token halting probability σ(w·h + b) → accumulate until ≥ threshold (1-ε) → remainder distributed → ponder cost = Σ(halting_prob) added to loss"
    invariant: "Every token halts within max_iterations. Ponder cost ≥ 1.0 per token. Halting probability monotonic non-decreasing."
    status: future

control:
  - advantage_id: om_recurrent_depth_execution
    verification: "Unit test: depth_iterations=3 → verify block called 3 times; verify same weights used each time"
    update_trigger: "Re-analyze when openmythos releases new version"
    last_verified: never

  - advantage_id: om_moe_routing
    verification: "Unit test: 4 experts, top-2 → verify exactly 2 experts selected per token; verify load balance loss ≥ 0"
    update_trigger: "Re-analyze when openmythos releases new version"
    last_verified: never

  - advantage_id: om_depth_lora_adaptation
    verification: "Unit test: depth 0 and depth 1 → verify different LoRA deltas applied; verify base weights identical"
    update_trigger: "Re-analyze when openmythos releases new version"
    last_verified: never

  - advantage_id: om_prelude_recurrent_coda
    verification: "Integration test: verify call order = prelude → recurrent × N → coda; skip prelude → verify error"
    update_trigger: "Re-analyze when openmythos releases new version"
    last_verified: never

  - advantage_id: om_convergence_stability_check
    verification: "Unit test: delta below threshold → verify early exit at correct iteration; delta above threshold → verify full depth_iterations"
    update_trigger: "Re-analyze when openmythos releases new version"
    last_verified: never

  - advantage_id: om_adaptive_computation_time
    verification: "Unit test: token with high complexity → verify more iterations; simple token → verify fewer iterations; all tokens → verify ponder cost ≥ 1.0"
    update_trigger: "Re-analyze when openmythos releases new version"
    last_verified: never
```
