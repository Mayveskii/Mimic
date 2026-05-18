# Git Domain — Sources

Where the git domain behavior comes from.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- BinaryRuntime with OpCodes: every git operation is a token (OpPacket), not arbitrary command.
- Tool loop: session, inference pool, hooks around git operations.
- Session tracking: git state tracked across operations, no stale assumptions.

**What Mimic does with them:**
Every git process is a chain of OpPackets with defined order and invariants. No shortcuts like `git commit -am`. Each step explicit.

**What Mimic does NOT copy:**
- Embryo's Go implementation details. Only the tokenization + deterministic chaining pattern.
- Specific opcode values from embryo (Mimic uses its own enumeration).

---

## bun (PR #30412)

**Principles taken:**
- Phase graph: git operations live inside CLASSIFY→PLAN→VALIDATE→EXEC→VERIFY→RESPOND.
- Never-rules: {git reset, git checkout for destructive purpose, git restore, git stash, git rebase} = hard deny.
- 2-vote verification: merge into production branch requires two independent verifications.
- Edit scope isolation: git operations on one repo do not interfere with operations on another.

**What Mimic does with them:**
Never-rules enforced before any git process. 2-vote on critical operations. Resource bitmask prevents cross-repo contamination.

**What Mimic does NOT copy:**
- Bun's crate-based scope isolation (Mimic uses domain/resource-based isolation).
- Bun's specific git wrapper API.

---

## gastown

**Principles taken:**
- Zero False Confidence: never trust cached git state — always verify with `git status`.
- Rollback on failure: if any git operation fails, all changes from current process are rolled back.
- Atomic allocation: branch names allocated with collision check (flock-like) to prevent race conditions.

**What Mimic does with them:**
Every git process starts with fresh status check. Failed process rolls back to pre-operation state. Branch creation checks for existing names.

**What Mimic does NOT copy:**
- Gastown's event-driven convoy system (Mimic uses orchestrator pipeline).
- Gastown's specific error classification.

---

## hermes-agent

**Principles taken:**
- Streaming health: git operations on remote repos have timeout (no infinite hangs).
- Error classification: git errors classified as retryable (network) vs permanent (conflict) vs auth.
- Tool guardrails: OP_FLAG_DANGEROUS on destructive git operations (force-push, reset) always blocked.

**What Mimic does with them:**
Git fetch/push have timeouts. Errors classified before retry decision. Dangerous flags require explicit model approval.

**What Mimic does NOT copy:**
- Hermes-agent's credential pool integration (Mimic uses its own credential system).
- Hermes-agent's specific retry backoff formula.

---

## binary-mesh / gonka (Anti-Patterns)

**Principles taken:**
- Context injection without structure stability (AP-05/AP-24): git operations must not mutate shared state without guaranteeing structure is stable.
- Flip-flop decisions (AP-16): never revert a git decision within minutes without evidence.
- Stale cached state (AP-11): never trust that "branch is up to date" without explicit fetch.
- buildMessages corruption: shared state mutation → message corruption.

**What Mimic does with them:**
Every git operation validates state from observable reality, not cache. Reverts are tracked as NEGATIVE artifacts with counter_pattern. Fetch is mandatory before merge decisions.

---

## Standard Git Practice

**Principles taken:**
- Fast-forward merges for clean history.
- Feature branch workflow for isolated changes.
- Descriptive commit messages.
- No commit without review.

**What Mimic does with them:**
FF-only enforced. Feature branch pattern tokenized. Commit message length validated. Diff review mandatory.
