# IO Domain — Sources

Where the IO domain behavior comes from.

---

## embryo (Mayveskii/embryo)

**Principles taken:**
- pkg/projectmap/: SQLite-based project navigation with FTS5 full-text search.
- Auto-index on WRITE: every file write triggers index update.
- Symbol lookup: imports, definitions, calls indexed for navigation.

**What Mimic does with them:**
Every file write triggers async index update. Symbol lookup via FTS5. Project map stays current.

**What Mimic does NOT copy:**
- Embryo's specific SQLite schema (Mimic's schema is spec-driven).
- Embryo's watcher-based indexing (Mimic uses explicit post-write trigger).

---

## caveman

**Principles taken:**
- Sensitive path protection: .env, credentials, keys → auto-exclude from read/write.
- File type detection: magic bytes + content, not just extension.

**What Mimic does with them:**
Sensitive paths blocked from read/write. File types detected for appropriate handling. Binary files not treated as text.

**What Mimic does NOT copy:**
- Caveman's specific path list (Mimic uses configurable pattern list).
- Caveman's ignore-file handling.

---

## gastown

**Principles taken:**
- Rollback on failure: file operations must be reversible.
- Best-effort cleanup: if operation fails, restore previous state.
- Atomic allocation: path existence checked before creation.

**What Mimic does with them:**
Write operations create backups before overwrite. Rollback restores from backup. Parent directories created atomically.

---

## rustnet

**Principles taken:**
- Sandbox: Landlock/Seatbelt prevents filesystem escape.
- Path validation: all paths within workspace boundary.

**What Mimic does with them:**
Workspace boundary enforced. Symlinks validated. Path traversal blocked.

---

## graphify

**Principles taken:**
- AST-based symbol extraction from source files.
- Two-pass analysis: structural + call-graph.

**What Mimic does with them:**
Index updates include AST extraction for symbol tables. Call-graph information added to project map.

---

## Anti-Patterns

**Principles taken from AP-04, AP-15, AP-29:**
- Validate input before I/O.
- Complete rollback, not partial.
- Surface all errors, never swallow.

**What Mimic does with them:**
IINV-01 validates existence. IINV-02 creates backups. IINV-05 verifies writes. All errors returned to model.
