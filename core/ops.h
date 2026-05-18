#ifndef MIMIC_OPS_H
#define MIMIC_OPS_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

/* ============================================================================
 * Constants
 * ============================================================================ */
#define MAX_OPS             1024
#define MAX_PACKET_SIZE     256
#define MAX_ARGS            16
#define OP_NAME_LEN         32
#define OP_DESC_LEN         128
#define MAX_OPEN_FDS        64
#define MAX_MMAP_REGIONS    64
#define ERR_MSG_LEN         256
#define MAX_BACKUP_PATH     512
#define CONFLICT_NONE       0
#define CONFLICT_LOW        1
#define CONFLICT_MEDIUM     2
#define CONFLICT_HIGH       3
#define CONFLICT_FATAL      4

/* ============================================================================
 * Flags
 * ============================================================================ */
#define OP_FLAG_SAFE        0x01
#define OP_FLAG_READONLY    0x02
#define OP_FLAG_ATOMIC      0x04
#define OP_FLAG_REVERSIBLE  0x08
#define OP_FLAG_NETWORK     0x10
#define OP_FLAG_DISK        0x20
#define OP_FLAG_MEMORY      0x40
#define OP_FLAG_DANGEROUS   0x80

/* ============================================================================
 * Error codes
 * ============================================================================ */
#define ERR_OK                   0
#define ERR_NOT_INIT             1
#define ERR_INVALID_CHAIN        1
#define ERR_INVALID_OPCODE       2
#define ERR_UNREGISTERED         3
#define ERR_BAD_ARGS             4
#define ERR_BAD_FD               5
#define ERR_ENERGY_OVERFLOW      6
#define ERR_PERMISSION           7
#define ERR_CONFLICT             8
#define ERR_ROLLBACK_FAIL        9
#define ERR_ATOMIC_BREAK        10
#define ERR_EXEC_FAIL           11
#define ERR_INVALID_RESULT      12
#define ERR_CIRCUIT_BROKEN      15

/* ============================================================================
 * OpCode Enumeration
 * ============================================================================ */
typedef enum {
    OP_NOP = 0,

    /* Memory (0x10-0x1F) */
    OP_MMAP_ALLOC   = 0x10,
    OP_MMAP_FREE    = 0x11,
    OP_MMAP_READ    = 0x12,
    OP_MMAP_WRITE   = 0x13,
    OP_MMAP_SYNC    = 0x14,

    /* I/O (0x20-0x2F) */
    OP_IO_READ      = 0x20,
    OP_IO_WRITE     = 0x21,
    OP_IO_OPEN      = 0x22,
    OP_IO_CLOSE     = 0x23,
    OP_IO_SEEK      = 0x24,

    /* Git (0x30-0x3D) */
    OP_GIT_INIT     = 0x30,
    OP_GIT_CLONE    = 0x31,
    OP_GIT_FETCH    = 0x32,
    OP_GIT_STATUS   = 0x33,
    OP_GIT_DIFF     = 0x34,
    OP_GIT_ADD      = 0x35,
    OP_GIT_COMMIT   = 0x36,
    OP_GIT_PUSH     = 0x37,
    OP_GIT_CHECKOUT = 0x38,
    OP_GIT_BRANCH   = 0x39,
    OP_GIT_MERGE    = 0x3A,
    OP_GIT_REBASE   = 0x3B,
    OP_GIT_TAG      = 0x3C,
    OP_GIT_RESET    = 0x3D,

    /* Build (0x40-0x4F) */
    OP_BUILD_COMPILE = 0x40,
    OP_BUILD_LINK    = 0x41,
    OP_BUILD_TEST    = 0x42,
    OP_BUILD_DEPLOY  = 0x43,
    OP_BUILD_CLEAN   = 0x44,

    /* Network (0x50-0x5F) */
    OP_NET_HTTP_GET     = 0x50,
    OP_NET_HTTP_POST    = 0x51,
    OP_NET_TCP_CONNECT  = 0x52,
    OP_NET_TCP_SEND     = 0x53,
    OP_NET_TCP_RECV     = 0x54,
    OP_NET_TCP_CLOSE    = 0x55,
    OP_NET_WEBSOCKET    = 0x56,

    /* Process (0x60-0x6F) */
    OP_PROC_SPAWN   = 0x60,
    OP_PROC_WAIT    = 0x61,
    OP_PROC_KILL    = 0x62,
    OP_PROC_SIGNAL  = 0x63,

    /* Utility (0x70-0x7F) */
    OP_HASH_SHA256  = 0x70,
    OP_HASH_MD5     = 0x71,
    OP_COMPRESS_GZIP = 0x72,
    OP_DECOMPRESS_GZIP = 0x73,
    OP_ENCRYPT_AES  = 0x74,
    OP_DECRYPT_AES  = 0x75,

    /* System (0x80-0x8F) */
    OP_SYS_EXEC         = 0x80,
    OP_SYS_ENV_GET      = 0x81,
    OP_SYS_ENV_SET      = 0x82,
    OP_SYS_FILE_EXISTS  = 0x83,
    OP_SYS_DIR_CREATE   = 0x84,
    OP_SYS_DIR_REMOVE   = 0x85,
    OP_SYS_FILE_COPY    = 0x86,
    OP_SYS_FILE_MOVE    = 0x87,
    OP_SYS_FILE_DELETE  = 0x88,
    OP_SYS_CHMOD        = 0x89,

    /* Session / Orchestrator (0x90-0x9A) */
    OP_SESS_BUDGET_CHECK    = 0x90,
    OP_SESS_CONTEXT_APPEND  = 0x91,
    OP_SESS_DENIAL_RECORD   = 0x92,
    OP_SESS_SNAPSHOT        = 0x93,
    OP_SESS_COMPRESS        = 0x94,
    OP_ORCH_CLASSIFY        = 0x95,
    OP_ORCH_PLAN            = 0x96,
    OP_ORCH_VALIDATE        = 0x97,
    OP_ORCH_EXEC            = 0x98,
    OP_ORCH_VERIFY          = 0x99,
    OP_ORCH_RESPOND         = 0x9A,

    /* Research (0xA0-0xAC) */
    OP_RESEARCH_HYPOTHESIS_CREATE     = 0xA0,
    OP_RESEARCH_HYPOTHESIS_LOAD       = 0xA1,
    OP_RESEARCH_HYPOTHESIS_INFERENCE  = 0xA2,
    OP_RESEARCH_EXPERIMENT_RUN        = 0xA3,
    OP_RESEARCH_RESULT_STORE          = 0xA4,
    OP_RESEARCH_STATISTICAL_TEST      = 0xA5,
    OP_RESEARCH_LITERATURE_FETCH      = 0xA6,
    OP_RESEARCH_LITERATURE_PARSE      = 0xA7,
    OP_RESEARCH_LITERATURE_INDEX      = 0xA8,
    OP_RESEARCH_CITATION_LINK         = 0xA9,
    OP_RESEARCH_LITERATURE_EMBED      = 0xAA,
    OP_RESEARCH_PROGRESS_STORE        = 0xAB,
    OP_RESEARCH_CONTEXT_SUMMARIZE     = 0xAC,

    /* Self-Management (0xB0-0xB5) */
    OP_SELF_CHECKPOINT_CREATE   = 0xB0,
    OP_SELF_CHECKPOINT_RESTORE  = 0xB1,
    OP_SELF_BUDGET_REALLOCATE   = 0xB2,
    OP_SELF_STRATEGY_PIVOT      = 0xB3,
    OP_SELF_PROGRESS_ASSESS     = 0xB4,
    OP_SELF_CONTEXT_SUMMARIZE   = 0xB5,

    OP_MAX = 0xFF
} OpCode;

/* ============================================================================
 * Argument types
 * ============================================================================ */
enum {
    ARG_TYPE_INT = 0,
    ARG_TYPE_FLOAT = 1,
    ARG_TYPE_STRING = 2,
    ARG_TYPE_BOOL = 3,
    ARG_TYPE_BLOB = 4,
};

typedef struct {
    char key[32];
    uint8_t type;
    union {
        int64_t i;
        double f;
        char s[256];
        bool b;
        struct { void *buffer; size_t buffer_size; } blob;
    } value;
} OpArg;

/* ============================================================================
 * Extended packet with arguments
 * ============================================================================ */
typedef struct {
    uint8_t opcode;
    uint8_t flags;
    uint16_t slot;
    uint32_t arg_count;
    OpArg args[MAX_ARGS];
    int fd_in;
    int fd_out;
    char result[4096];  /* Output buffer for operation results */
    size_t result_len;
} OpPacketEx;

/* ============================================================================
 * OpCode definition (registry entry)
 * ============================================================================ */
typedef struct {
    OpCode opcode;
    char name[OP_NAME_LEN];
    char description[OP_DESC_LEN];
    int (*execute)(OpPacketEx* packet);
    OpCode inverse_opcode;
    int (*inverse_execute)(OpPacketEx* original, OpPacketEx* inverse_result);
    uint32_t required_flags;
    uint32_t forbidden_flags;
    float cost_tokens;
    float cost_time_us;
    float cost_memory_bytes;
    uint8_t safety_level;
    bool is_reversible;
    bool is_atomic;
} OpCodeDef;

/* ============================================================================
 * Validation result
 * ============================================================================ */
typedef struct {
    bool is_valid;
    uint32_t error_code;
    char error_msg[ERR_MSG_LEN];
    uint32_t invalid_op_index;
    uint32_t conflict_op_pair[2];
    float total_energy;
    float estimated_latency_us;
} ValidationResult;

/* ============================================================================
 * Execution context
 * ============================================================================ */
typedef struct {
    int open_fds[MAX_OPEN_FDS];
    size_t fd_count;
    void* mmap_regions[MAX_MMAP_REGIONS];
    size_t mmap_sizes[MAX_MMAP_REGIONS];
    size_t mmap_count;
    uint64_t pre_state_hash;
    void* pre_state_blob;
    size_t pre_state_blob_size;
    float session_budget_tokens;
    float session_budget_time_ms;
    bool circuit_broken;
    uint32_t denial_count;
    uint32_t chain_id;
} ExecContext;

/* ============================================================================
 * Lifecycle
 * ============================================================================ */
int ops_init(void);
void ops_shutdown(void);
int ops_register(OpCodeDef* def);
const OpCodeDef* ops_get_definition(OpCode opcode);
void ops_register_builtins(void);

/* ============================================================================
 * Execution
 * ============================================================================ */
int ops_execute(OpPacketEx* packet);
int ops_execute_chain(OpPacketEx* packets, size_t count, ExecContext* ctx);
ValidationResult ops_validate_chain(const OpPacketEx* packets, size_t count, ExecContext* ctx);
bool ops_check_conflict(OpCode a, OpCode b);
float ops_calculate_action(const OpPacketEx* packets, size_t count);

/* ============================================================================
 * Utilities
 * ============================================================================ */
const char* ops_opcode_to_string(OpCode opcode);
OpCode ops_string_to_opcode(const char* name);
void ops_packet_init(OpPacketEx* packet, OpCode opcode);
void ops_packet_set_string(OpPacketEx* packet, const char* key, const char* value);
void ops_packet_set_int(OpPacketEx* packet, const char* key, int64_t value);

/* ============================================================================
 * Memory
 * ============================================================================ */
void* ops_mmap_alloc(size_t size);
int ops_mmap_free(void* ptr, size_t size);
int ops_mmap_sync(void* ptr, size_t size);

/* ============================================================================
 * Time
 * ============================================================================ */
uint64_t ops_get_time_ns(void);

/* ============================================================================
 * Rollback
 * ============================================================================ */
int ops_rollback_chain(OpPacketEx* packets, uint32_t failed_index, ExecContext* ctx);
int ops_create_backup(const char* path, char* backup_path, size_t backup_path_size);
void ops_best_effort_cleanup(OpPacketEx* packet, ExecContext* ctx);

/* ============================================================================
 * State
 * ============================================================================ */
uint64_t ops_compute_state_hash(ExecContext* ctx);

/* ============================================================================
 * Permission stubs (to be connected to session layer)
 * ============================================================================ */
bool session_has_explicit_allow(OpCode opcode, uint32_t chain_id);
bool session_has_2vote_verify(uint32_t chain_id);

#endif /* MIMIC_OPS_H */
