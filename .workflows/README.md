# Mimic Workflows

Predefined agent execution patterns. Each workflow is a JSON file that defines an ordered sequence of tool calls with parameters, preconditions, and verification steps.

## Usage

When an agent receives a task, it should:
1. **Classify** the task against workflow names
2. **Load** the matching workflow JSON
3. **Execute** steps sequentially, substituting `{{parameters}}`
4. **Verify** assertions at each step
5. **Report** results

## Available Workflows

| Workflow | Purpose | Steps | Tools Used |
|----------|---------|-------|------------|
| `fix-typo.json` | Fix typo and verify | 3 | SYS_FILE_READ, FILE_EDIT |
| `read-and-edit.json` | General file editing | 4 | SYS_FILE_READ, FILE_EDIT |
| `git-atomic-commit.json` | Safe git commit | 4 | GIT_STATUS, GIT_DIFF, GIT_ADD, GIT_COMMIT |
| `build-and-test.json` | Compile and verify | 2 | BUILD_COMPILE, BUILD_TEST |
| `explore-codebase.json` | Gather context | 3 | SYS_FILE_READ, SYS_FILE_EXISTS |

## Workflow Schema

```json
{
  "name": "workflow-name",
  "version": "1.0",
  "description": "What this does",
  "intent": "When to use this",
  "steps": [
    {
      "step": 1,
      "tool": "TOOL_NAME",
      "description": "Human-readable purpose",
      "arguments": { "param": "{{value}}" },
      "save_result_as": "variable_name",
      "condition": "optional guard",
      "optional": false,
      "assert": "post-condition to verify"
    }
  ],
  "parameters": {
    "param_name": {
      "type": "string",
      "required": true,
      "description": "What this parameter means"
    }
  },
  "invariants": [ "conditions that must hold" ],
  "rollback": { "on_step_failure": "recovery_action" },
  "energy_estimate": { "tokens": 10, "latency_us": 20000 }
}
```

## Adding New Workflows

1. Copy an existing workflow as template
2. Define `steps` with tool names from `internal/mcp/tool_schemas.go`
3. Specify `parameters` with types and defaults
4. Add `invariants` for safety checks
5. Test with both qwen and kimi models
6. Submit PR to `.workflows/` directory

## Integration with Orchestrator

The orchestrator (`internal/orchestrator/`) loads workflows from this directory at startup. During task classification, it matches agent intent against workflow `intent` fields using keyword overlap.

## Example Execution

**Task:** "Fix typo 'recieve' -> 'receive' in README.md"

```
1. Classify → matches "fix-typo"
2. Load workflow → fix-typo.json
3. Substitute parameters:
   - {{file_path}} = "README.md"
   - {{typo}} = "recieve"
   - {{correction}} = "receive"
4. Execute steps 1-3
5. Verify step 3 assertion passes
6. Report success
```

## TODO

- [ ] Add `safe-refactor.json` — rename variable across files
- [ ] Add `add-feature.json` — implement + test + commit
- [ ] Add `code-review.json` — read → analyze → comment
- [ ] Add `batch-process.json` — loop over files
- [ ] Workflow validation tool (check all tool names exist)
