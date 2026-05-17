# SEMANTICS.md — Mimic

Every function: name | input | output | invariant | source

---

## C-Core: ops.c

| Function | Input | Output | Invariant | Source |
|----------|-------|--------|-----------|--------|
| ops_init | void | int (0=ok) | Cannot init twice | ops.c |
| ops_shutdown | void | void | Only after init | ops.c |
| ops_register | OpCodeDef* | int (0=ok) | opcode < OP_MAX, registry not full | ops.c |
| ops_get_definition | OpCode | const OpCodeDef* | NULL if opcode ≥ OP_MAX | ops.c |
| ops_execute | OpPacket* | int (0=ok, -1=invalid, -2=bad opcode, -3=no executor) | Measures latency_ns via CLOCK_MONOTONIC | ops.c |
| ops_execute_chain | OpPacket[], count, ExecContext* | int (0=ok, -4=validation fail) | validate_chain passed BEFORE first exec | ops.c |
| ops_validate_chain | OpPacket[], count | ValidationResult | O(n²) pairwise conflict check | ops.c |
| ops_check_conflict | OpCode, OpCode | bool | invalid opcodes → conflict | ops.c |
| ops_calculate_action | OpPacket[], count | float | S = Σ(cost_tokens × cost_time_us) | ops.c |
| ops_get_time_ns | void | uint64_t | CLOCK_MONOTONIC | ops.c |
| ops_opcode_to_string | OpCode | const char* | "UNKNOWN" if opcode ≥ OP_MAX | ops.c |
| ops_string_to_opcode | const char* | OpCode | OP_NOP if not found | ops.c |
| ops_packet_init | OpPacket*, OpCode | void | id auto-increment, all zeroed | ops.c |
| ops_packet_set_string | OpPacket*, key, value | void | arg_count < MAX_ARGS=16 | ops.c |
| ops_packet_set_int | OpPacket*, key, int | void | arg_count < MAX_ARGS=16 | ops.c |
| ops_mmap_alloc | size_t | void* (NULL=fail) | MAP_PRIVATE|MAP_ANONYMOUS | ops.c |
| ops_mmap_free | void*, size_t | int (0=ok) | ptr ≠ NULL, size ≠ 0 | ops.c |
| ops_mmap_sync | void*, size_t | int (0=ok) | MS_SYNC | ops.c |
| ops_register_builtins | void | void | NOP + SYS_FILE_EXISTS + SYS_DIR_CREATE | ops.c |

## C-Core: mmap_ops.c

| Function | Input | Output | Invariant | Source |
|----------|-------|--------|-----------|--------|
| mmap_ops_register_all | void | void | Registers OP_MMAP_* executors | mmap_ops.c |

## C-Core: git_ops.c

| Function | Input | Output | Invariant | Source |
|----------|-------|--------|-----------|--------|
| git_ops_register_all | void | void | Registers OP_GIT_* executors | git_ops.c |

## C-Core: git_scenarios.c

| Function | Input | Output | Invariant | Source |
|----------|-------|--------|-----------|--------|
| scenario_atomic_commit | path, message | ScenarioResult | status+diff+commit atomically | git_scenarios.c |
| scenario_safe_merge | source, target | ScenarioResult | Fast-forward only | git_scenarios.c |
| scenario_feature_branch | name | ScenarioResult | Create branch without switching | git_scenarios.c |
| scenario_hotfix | name, target | ScenarioResult | branch + commit + merge into target | git_scenarios.c |
| scenario_ci_diff_check | base, head | ScenarioResult | diff --check, no whitespace errors | git_scenarios.c |

## C-Core: git_search_ops.c

| Function | Input | Output | Invariant | Source |
|----------|-------|--------|-----------|--------|
| (registration functions) | void | void | ⚠️ Opcode collision 0x50-0x5F with NET | git_search_ops.c — NEEDS FIX |

## C-Core: libbmap.a (no .c sources)

| Function | Input | Output | Invariant | Source |
|----------|-------|--------|-----------|--------|
| bmap_open | path | bmap_t* | NULL if not exists | libbmap.a ⚠️ no .c |
| bmap_close | bmap_t* | void | — | libbmap.a ⚠️ no .c |
| bmap_read_cell | bmap_t*, cell_id | cell_data | — | libbmap.a ⚠️ no .c |
| bmap_write_cell | bmap_t*, cell_id, data | int | — | libbmap.a ⚠️ no .c |
| bmap_write | bmap_t* | int | — | libbmap.a ⚠️ no .c |
| bmap_free_cell | bmap_t*, cell_id | int | — | libbmap.a ⚠️ no .c |
| bmap_cell_serialized_size | cell | size_t | — | libbmap.a ⚠️ no .c |
| si_create | void | slot_index_t* | — | libbmap.a ⚠️ no .c |
| si_destroy | slot_index_t* | void | — | libbmap.a ⚠️ no .c |
| si_insert | slot_index_t*, slot | int | — | libbmap.a ⚠️ no .c |
| si_query_domain | slot_index_t*, domain | result_set | — | libbmap.a ⚠️ no .c |
| si_query_domain_layer | slot_index_t*, domain, layer | result_set | — | libbmap.a ⚠️ no .c |
| si_query_state_hash | slot_index_t*, hash | result_set | — | libbmap.a ⚠️ no .c |
| si_build_from_bmap | bmap_t* | slot_index_t* | — | libbmap.a ⚠️ no .c |
| si_result_free | result_set | void | — | libbmap.a ⚠️ no .c |
| inv_create | void | invariant_t* | — | libbmap.a ⚠️ no .c |
| inv_destroy | invariant_t* | void | — | libbmap.a ⚠️ no .c |
| inv_add | invariant_t*, condition | int | — | libbmap.a ⚠️ no .c |
| inv_find_similar | invariant_t*, condition, threshold | result_set | — | libbmap.a ⚠️ no .c |
| inv_load | path | invariant_t* | — | libbmap.a ⚠️ no .c |
| inv_save | invariant_t*, path | int | — | libbmap.a ⚠️ no .c |
| inv_dedup_check | invariant_t*, condition | bool | — | libbmap.a ⚠️ no .c |
| gnk_compute | bmap_t*, domain | gnk_result | — | libbmap.a ⚠️ no .c |
| gnk_score_domains | bmap_t* | gnk_result* | — | libbmap.a ⚠️ no .c |
| gnk_result_free | gnk_result* | void | — | libbmap.a ⚠️ no .c |
| snapshot_build | bmap_t* | snapshot_t* | — | libbmap.a ⚠️ no .c |
| snapshot_load | path | snapshot_t* | — | libbmap.a ⚠️ no .c |
| snapshot_write | snapshot_t*, path | int | — | libbmap.a ⚠️ no .c |
| snapshot_sign | snapshot_t*, key | int | — | libbmap.a ⚠️ no .c |
| snapshot_diff | snapshot_t*, snapshot_t* | diff_result | — | libbmap.a ⚠️ no .c |
| snapshot_diff_free | diff_result | void | — | libbmap.a ⚠️ no .c |
| snapshot_free | snapshot_t* | void | — | libbmap.a ⚠️ no .c |
| layer_walk | bmap_t*, layer | walk_result | — | libbmap.a ⚠️ no .c |
| drift_detect | bmap_t*, snapshot_t* | drift_result | — | libbmap.a ⚠️ no .c |
| cosine_f32 | float[], float[], n | float | [-1, 1] | libbmap.a ⚠️ no .c |
| cosine_int8 | int8[], int8[], n | float | [-1, 1] | libbmap.a ⚠️ no .c |
| batch_cosine_int8 | int8[][], int8[], n, batch | float[] | [-1, 1] each | libbmap.a ⚠️ no .c |
| int8_quantize | float[], n | int8[] | — | libbmap.a ⚠️ no .c |
| sha256_hash | data, len | hash[32] | — | libbmap.a ⚠️ no .c |
| z_density_compute | bmap_t*, domain | float | ≥ 0 | libbmap.a ⚠️ no .c |

## CGO Bridge: cgo_wrapper.go

| Function | Input | Output | Invariant | Source |
|----------|-------|--------|-----------|--------|
| Init | void | error | Calls ops_init + ops_register_builtins | cgo_wrapper.go |
| Shutdown | void | void | Calls ops_shutdown | cgo_wrapper.go |
| Execute | *OpPacket | error | RLock → toCPacket → ops_execute → fromCPacket → freeCPacket | cgo_wrapper.go |
| ExecuteChain | []*OpPacket, *ExecContext | error | validate_chain embedded in ops_execute_chain | cgo_wrapper.go |
| ValidateChain | []*OpPacket | *ValidationResult, error | Go-side + C-side validator | cgo_wrapper.go |
| CheckConflict | OpCode, OpCode | bool | Delegates to ops_check_conflict | cgo_wrapper.go |
| GetTimeNs | void | uint64 | Delegates to ops_get_time_ns | cgo_wrapper.go |
| CalculateAction | []*OpPacket | float32 | Delegates to ops_calculate_action | cgo_wrapper.go |
| MMapAlloc | int | []byte, error | unsafe.Slice on C memory | cgo_wrapper.go |
| MMapFree | []byte | error | Delegates to ops_mmap_free | cgo_wrapper.go |
| MMapSync | []byte | error | Delegates to ops_mmap_sync | cgo_wrapper.go |
| toCPacket | *OpPacket | *C.OpPacket | malloc, memcpy, C strings freed after copy | cgo_wrapper.go |
| fromCPacket | *C.OpPacket, *OpPacket | void | Copies result back | cgo_wrapper.go |
| freeCPacket | *C.OpPacket | void | Frees buffer + struct | cgo_wrapper.go |

## CGO Bridge: validator.go

| Function | Input | Output | Invariant | Source |
|----------|-------|--------|-----------|--------|
| NewValidator | void | *Validator | Default rules + conflictRules | validator.go |
| AddRule | name, ValidationRule | void | — | validator.go |
| AddConflictRule | name, ConflictRule | void | — | validator.go |
| ValidateOne | *OpPacket | error | All rules sequentially | validator.go |
| ValidateChain | []*OpPacket | *ValidationResult | MaxChainLength=1024, MaxTotalBuffer=10MB | validator.go |
| GetOpDefinition | OpCode | *OpCodeDef | Stub: only 3 explicit, rest default | validator.go ⚠️ |

## CGO Bridge: helpers.go

| Function | Input | Output | Invariant | Source |
|----------|-------|--------|-----------|--------|
| NewOpPacket | OpCode, ...OpArg | *OpPacket | OP_FLAG_SAFE by default | helpers.go |
| NewOpPacketWithBuffer | OpCode, []byte, ...OpArg | *OpPacket | OP_FLAG_SAFE by default | helpers.go |
| WithFlag | *OpPacket, uint32 | *OpPacket | builder pattern | helpers.go |
| WithTimeout | *OpPacket, uint32 | *OpPacket | builder pattern | helpers.go |
| WithRetry | *OpPacket, uint32 | *OpPacket | builder pattern | helpers.go |
| WithArg | *OpPacket, key, value, type | *OpPacket | builder pattern | helpers.go |
| WithIntArg | *OpPacket, key, int | *OpPacket | builder pattern | helpers.go |
| WithStringArg | *OpPacket, key, value | *OpPacket | builder pattern | helpers.go |

## TODO (functions without implementation)

These OpCodes are declared in ops.h but have no executors:

OP_IO_READ, OP_IO_WRITE, OP_IO_OPEN, OP_IO_CLOSE, OP_IO_SEEK,
OP_BUILD_COMPILE, OP_BUILD_LINK, OP_BUILD_TEST, OP_BUILD_DEPLOY, OP_BUILD_CLEAN,
OP_NET_HTTP_GET, OP_NET_HTTP_POST, OP_NET_TCP_CONNECT, OP_NET_TCP_SEND, OP_NET_TCP_RECV, OP_NET_TCP_CLOSE,
OP_PROC_SPAWN, OP_PROC_WAIT, OP_PROC_KILL, OP_PROC_SIGNAL,
OP_HASH_SHA256, OP_HASH_MD5, OP_COMPRESS_GZIP, OP_DECOMPRESS_GZIP, OP_ENCRYPT_AES, OP_DECRYPT_AES,
OP_SYS_EXEC, OP_SYS_ENV_GET, OP_SYS_ENV_SET, OP_SYS_DIR_REMOVE, OP_SYS_FILE_COPY, OP_SYS_FILE_MOVE, OP_SYS_FILE_DELETE

⚠️ libbmap.a: 39 functions without .c sources — require rewrite
⚠️ GetOpDefinition in validator.go: stub, only 3 explicit definitions
