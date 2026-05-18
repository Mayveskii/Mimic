# Pattern: build_and_test

Tokenized process for compiling and testing code.

---

## When to Use

Model wants to verify code compiles and passes tests. Intent: "build this project", "run tests", "check if code works".

## Goal

Code compiles with zero errors. Tests pass (or failures reported). Compilation failures caught before test execution. Test timeout enforced.

## Chain (Tokenized)

```c
OpPacket packets[] = {
    {OP_SYS_FILE_EXISTS,  .args = {{"path", "go.mod"}}},  // Detect build system
    {OP_BUILD_COMPILE,    .args = {{"target", "./..."}}},
    {OP_BUILD_TEST,       .args = {{"filter", ""}, {"timeout_ms", "300000"}}}
};
```

### Step-by-Step Semantics

**Step 1: Detect Build System**
- Check for `go.mod` → Go project.
- Check for `Cargo.toml` → Rust project.
- Check for `package.json` → Node project.
- Check for `Makefile` → Make project.
- If multiple found → prefer in order: Makefile > language-specific > generic.

**Step 2: Compile**
- Run appropriate compile command based on detected build system.
- Capture stdout/stderr.
- If compile errors → STOP. Return error output to model. DO NOT run tests.

**Step 3: Test**
- If compile succeeds → run tests.
- Apply timeout (default 300s, configurable).
- Capture: pass count, fail count, skip count, coverage.
- If timeout → `ERR_TIMEOUT`, partial results returned.

## Hard Constraints

- Tests NEVER run if compilation fails.
- Compilation must produce zero errors. Warnings logged but do not block.
- Test timeout MUST be set and enforced.
- If build system unknown → ask model for clarification.

## Invariants

- compile_success → test_may_run.
- compile_failure → test_never_runs.
- All test failures include file, line, and function name.
- Working tree unchanged by build (output goes to build dir).

## Result When Successful

```json
{
  "status": "success",
  "compile": {
    "command": "go build ./...",
    "duration_ms": 12400,
    "errors": 0,
    "warnings": 3
  },
  "test": {
    "command": "go test ./...",
    "duration_ms": 45800,
    "passed": 47,
    "failed": 0,
    "skipped": 2,
    "coverage": 0.73
  }
}
```

## Result When Failed

```json
{
  "status": "failure",
  "stage": "compile",
  "error": "undefined: SomeFunc in main.go:42",
  "compile_output": "...full compiler output...",
  "test_not_run": true
}
```

## How a Model Uses This

Model says "build and test" → Mimic detects build system → compiles → if success → tests → returns results. If compile fails → model gets error output. Model never sees test failures for code that doesn't compile.

## Energy Cost (Estimated)

- File exists: tokens=1, latency=1us → energy=0.001
- Compile: tokens=6, latency=1000ms → energy=6000
- Test: tokens=6, latency=5000ms → energy=30000
- Total: ~36000 token-ms

## QAC Mapping

- QAC-3 (Determinism): Same code → same compile/test result.
- QAC-4 (State Validation): Compile validates syntax before test.
- QAC-6 (Timeout): Test timeout enforced.
- QAC-9 (Wholeness): All tests in scope run, not subset.
