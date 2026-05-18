#include "ops.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <time.h>
#include <sys/mman.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <dirent.h>
#include <libgen.h>
#include <stdint.h>
#include <signal.h>
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wdeprecated-declarations"
#include <openssl/sha.h>
#include <openssl/md5.h>
#pragma GCC diagnostic pop
#include <sys/wait.h>

/* ============================================================================
 * Static globals
 * ============================================================================ */
static bool g_initialized = false;
static OpCodeDef g_op_registry[MAX_OPS];
static uint8_t g_conflict_matrix[256][256];
static float g_energy_costs[256][3];

/* ============================================================================
 * Internal helpers
 * ============================================================================ */
static void init_conflict_matrix(void) {
    memset(g_conflict_matrix, CONFLICT_NONE, sizeof(g_conflict_matrix));

    /* Self-conflicts for dangerous / singleton operations */
    g_conflict_matrix[OP_SYS_EXEC][OP_SYS_EXEC]         = CONFLICT_FATAL;
    g_conflict_matrix[OP_SYS_FILE_DELETE][OP_SYS_FILE_DELETE] = CONFLICT_HIGH;
    g_conflict_matrix[OP_SYS_DIR_REMOVE][OP_SYS_DIR_REMOVE]   = CONFLICT_HIGH;
    g_conflict_matrix[OP_BUILD_DEPLOY][OP_BUILD_DEPLOY]       = CONFLICT_FATAL;
    g_conflict_matrix[OP_BUILD_CLEAN][OP_BUILD_CLEAN]         = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_COMMIT][OP_GIT_COMMIT]         = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_PUSH][OP_GIT_PUSH]               = CONFLICT_FATAL;
    g_conflict_matrix[OP_GIT_MERGE][OP_GIT_MERGE]             = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_REBASE][OP_GIT_REBASE]           = CONFLICT_HIGH;
    g_conflict_matrix[OP_PROC_SPAWN][OP_PROC_SPAWN]           = CONFLICT_MEDIUM;
    g_conflict_matrix[OP_PROC_KILL][OP_PROC_KILL]             = CONFLICT_HIGH;

    /* Cross-domain conflicts */
    g_conflict_matrix[OP_SYS_EXEC][OP_SYS_FILE_DELETE]        = CONFLICT_HIGH;
    g_conflict_matrix[OP_SYS_FILE_DELETE][OP_SYS_EXEC]        = CONFLICT_HIGH;
    g_conflict_matrix[OP_SYS_EXEC][OP_SYS_DIR_REMOVE]         = CONFLICT_HIGH;
    g_conflict_matrix[OP_SYS_DIR_REMOVE][OP_SYS_EXEC]           = CONFLICT_HIGH;
    g_conflict_matrix[OP_IO_WRITE][OP_SYS_FILE_DELETE]        = CONFLICT_HIGH;
    g_conflict_matrix[OP_SYS_FILE_DELETE][OP_IO_WRITE]          = CONFLICT_HIGH;
    g_conflict_matrix[OP_BUILD_CLEAN][OP_BUILD_COMPILE]       = CONFLICT_HIGH;
    g_conflict_matrix[OP_BUILD_COMPILE][OP_BUILD_CLEAN]       = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_COMMIT][OP_GIT_CHECKOUT]         = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_CHECKOUT][OP_GIT_COMMIT]         = CONFLICT_HIGH;
    g_conflict_matrix[OP_GIT_PUSH][OP_GIT_REBASE]             = CONFLICT_FATAL;
    g_conflict_matrix[OP_GIT_REBASE][OP_GIT_PUSH]             = CONFLICT_FATAL;
    g_conflict_matrix[OP_PROC_SPAWN][OP_PROC_KILL]            = CONFLICT_HIGH;
    g_conflict_matrix[OP_PROC_KILL][OP_PROC_SPAWN]              = CONFLICT_HIGH;
    g_conflict_matrix[OP_SYS_ENV_SET][OP_SYS_EXEC]              = CONFLICT_MEDIUM;
    g_conflict_matrix[OP_SYS_EXEC][OP_SYS_ENV_SET]              = CONFLICT_MEDIUM;
}

static void init_energy_costs(void) {
    memset(g_energy_costs, 0, sizeof(g_energy_costs));

    /* NOP */
    g_energy_costs[OP_NOP][0] = 0.0f;   g_energy_costs[OP_NOP][1] = 0.01f;   g_energy_costs[OP_NOP][2] = 0.0f;

    /* Memory */
    g_energy_costs[OP_MMAP_ALLOC][0] = 2.0f;  g_energy_costs[OP_MMAP_ALLOC][1] = 5.0f;  g_energy_costs[OP_MMAP_ALLOC][2] = 4096.0f;
    g_energy_costs[OP_MMAP_FREE][0] = 1.0f;   g_energy_costs[OP_MMAP_FREE][1] = 2.0f;   g_energy_costs[OP_MMAP_FREE][2] = 0.0f;
    g_energy_costs[OP_MMAP_READ][0] = 1.0f;   g_energy_costs[OP_MMAP_READ][1] = 1.0f;   g_energy_costs[OP_MMAP_READ][2] = 0.0f;
    g_energy_costs[OP_MMAP_WRITE][0] = 2.0f;  g_energy_costs[OP_MMAP_WRITE][1] = 2.0f;  g_energy_costs[OP_MMAP_WRITE][2] = 0.0f;
    g_energy_costs[OP_MMAP_SYNC][0] = 1.0f;   g_energy_costs[OP_MMAP_SYNC][1] = 10.0f;  g_energy_costs[OP_MMAP_SYNC][2] = 0.0f;

    /* I/O */
    g_energy_costs[OP_IO_READ][0] = 1.0f;   g_energy_costs[OP_IO_READ][1] = 5.0f;   g_energy_costs[OP_IO_READ][2] = 0.0f;
    g_energy_costs[OP_IO_WRITE][0] = 2.0f;  g_energy_costs[OP_IO_WRITE][1] = 10.0f;  g_energy_costs[OP_IO_WRITE][2] = 0.0f;
    g_energy_costs[OP_IO_OPEN][0] = 2.0f;   g_energy_costs[OP_IO_OPEN][1] = 50.0f;  g_energy_costs[OP_IO_OPEN][2] = 0.0f;
    g_energy_costs[OP_IO_CLOSE][0] = 1.0f;  g_energy_costs[OP_IO_CLOSE][1] = 5.0f;  g_energy_costs[OP_IO_CLOSE][2] = 0.0f;
    g_energy_costs[OP_IO_SEEK][0] = 1.0f;   g_energy_costs[OP_IO_SEEK][1] = 1.0f;  g_energy_costs[OP_IO_SEEK][2] = 0.0f;

    /* System */
    g_energy_costs[OP_SYS_FILE_EXISTS][0] = 1.0f;  g_energy_costs[OP_SYS_FILE_EXISTS][1] = 10.0f;  g_energy_costs[OP_SYS_FILE_EXISTS][2] = 0.0f;
    g_energy_costs[OP_SYS_DIR_CREATE][0] = 2.0f;   g_energy_costs[OP_SYS_DIR_CREATE][1] = 50.0f;   g_energy_costs[OP_SYS_DIR_CREATE][2] = 4096.0f;
    g_energy_costs[OP_SYS_DIR_REMOVE][0] = 2.0f;   g_energy_costs[OP_SYS_DIR_REMOVE][1] = 100.0f;  g_energy_costs[OP_SYS_DIR_REMOVE][2] = 0.0f;
    g_energy_costs[OP_SYS_FILE_COPY][0] = 3.0f;    g_energy_costs[OP_SYS_FILE_COPY][1] = 500.0f;   g_energy_costs[OP_SYS_FILE_COPY][2] = 0.0f;
    g_energy_costs[OP_SYS_FILE_MOVE][0] = 3.0f;     g_energy_costs[OP_SYS_FILE_MOVE][1] = 200.0f;   g_energy_costs[OP_SYS_FILE_MOVE][2] = 0.0f;
    g_energy_costs[OP_SYS_FILE_DELETE][0] = 1.0f;   g_energy_costs[OP_SYS_FILE_DELETE][1] = 20.0f;   g_energy_costs[OP_SYS_FILE_DELETE][2] = 0.0f;
    g_energy_costs[OP_SYS_CHMOD][0] = 1.0f;        g_energy_costs[OP_SYS_CHMOD][1] = 10.0f;         g_energy_costs[OP_SYS_CHMOD][2] = 0.0f;
    g_energy_costs[OP_SYS_ENV_GET][0] = 1.0f;       g_energy_costs[OP_SYS_ENV_GET][1] = 1.0f;        g_energy_costs[OP_SYS_ENV_GET][2] = 0.0f;
    g_energy_costs[OP_SYS_ENV_SET][0] = 1.0f;       g_energy_costs[OP_SYS_ENV_SET][1] = 2.0f;        g_energy_costs[OP_SYS_ENV_SET][2] = 0.0f;
    g_energy_costs[OP_SYS_EXEC][0] = 5.0f;          g_energy_costs[OP_SYS_EXEC][1] = 10000.0f;      g_energy_costs[OP_SYS_EXEC][2] = 0.0f;

    /* Build */
    g_energy_costs[OP_BUILD_COMPILE][0] = 10.0f;  g_energy_costs[OP_BUILD_COMPILE][1] = 500000.0f;  g_energy_costs[OP_BUILD_COMPILE][2] = 0.0f;
    g_energy_costs[OP_BUILD_LINK][0] = 5.0f;     g_energy_costs[OP_BUILD_LINK][1] = 100000.0f;     g_energy_costs[OP_BUILD_LINK][2] = 0.0f;
    g_energy_costs[OP_BUILD_TEST][0] = 10.0f;    g_energy_costs[OP_BUILD_TEST][1] = 300000.0f;     g_energy_costs[OP_BUILD_TEST][2] = 0.0f;
    g_energy_costs[OP_BUILD_DEPLOY][0] = 5.0f;   g_energy_costs[OP_BUILD_DEPLOY][1] = 600000.0f;  g_energy_costs[OP_BUILD_DEPLOY][2] = 0.0f;
    g_energy_costs[OP_BUILD_CLEAN][0] = 3.0f;    g_energy_costs[OP_BUILD_CLEAN][1] = 50000.0f;     g_energy_costs[OP_BUILD_CLEAN][2] = 0.0f;

    /* Git */
    g_energy_costs[OP_GIT_STATUS][0] = 2.0f;   g_energy_costs[OP_GIT_STATUS][1] = 100.0f;  g_energy_costs[OP_GIT_STATUS][2] = 0.0f;
    g_energy_costs[OP_GIT_DIFF][0] = 3.0f;    g_energy_costs[OP_GIT_DIFF][1] = 200.0f;   g_energy_costs[OP_GIT_DIFF][2] = 0.0f;
    g_energy_costs[OP_GIT_ADD][0] = 2.0f;     g_energy_costs[OP_GIT_ADD][1] = 50.0f;     g_energy_costs[OP_GIT_ADD][2] = 0.0f;
    g_energy_costs[OP_GIT_COMMIT][0] = 3.0f;   g_energy_costs[OP_GIT_COMMIT][1] = 100.0f;  g_energy_costs[OP_GIT_COMMIT][2] = 0.0f;
    g_energy_costs[OP_GIT_PUSH][0] = 5.0f;    g_energy_costs[OP_GIT_PUSH][1] = 5000.0f;   g_energy_costs[OP_GIT_PUSH][2] = 0.0f;
    g_energy_costs[OP_GIT_CHECKOUT][0] = 3.0f; g_energy_costs[OP_GIT_CHECKOUT][1] = 200.0f; g_energy_costs[OP_GIT_CHECKOUT][2] = 0.0f;
    g_energy_costs[OP_GIT_CLONE][0] = 10.0f;  g_energy_costs[OP_GIT_CLONE][1] = 30000.0f;  g_energy_costs[OP_GIT_CLONE][2] = 0.0f;

    /* Network */
    g_energy_costs[OP_NET_HTTP_GET][0] = 5.0f;  g_energy_costs[OP_NET_HTTP_GET][1] = 5000.0f;  g_energy_costs[OP_NET_HTTP_GET][2] = 0.0f;
    g_energy_costs[OP_NET_HTTP_POST][0] = 5.0f; g_energy_costs[OP_NET_HTTP_POST][1] = 5000.0f; g_energy_costs[OP_NET_HTTP_POST][2] = 0.0f;
    g_energy_costs[OP_NET_TCP_CONNECT][0] = 3.0f; g_energy_costs[OP_NET_TCP_CONNECT][1] = 2000.0f; g_energy_costs[OP_NET_TCP_CONNECT][2] = 0.0f;
    g_energy_costs[OP_NET_TCP_SEND][0] = 2.0f;  g_energy_costs[OP_NET_TCP_SEND][1] = 1000.0f;  g_energy_costs[OP_NET_TCP_SEND][2] = 0.0f;
    g_energy_costs[OP_NET_TCP_RECV][0] = 2.0f;  g_energy_costs[OP_NET_TCP_RECV][1] = 1000.0f;  g_energy_costs[OP_NET_TCP_RECV][2] = 0.0f;
    g_energy_costs[OP_NET_TCP_CLOSE][0] = 1.0f; g_energy_costs[OP_NET_TCP_CLOSE][1] = 50.0f;   g_energy_costs[OP_NET_TCP_CLOSE][2] = 0.0f;
    g_energy_costs[OP_NET_WEBSOCKET][0] = 5.0f; g_energy_costs[OP_NET_WEBSOCKET][1] = 3000.0f; g_energy_costs[OP_NET_WEBSOCKET][2] = 0.0f;

    /* Process */
    g_energy_costs[OP_PROC_SPAWN][0] = 5.0f;  g_energy_costs[OP_PROC_SPAWN][1] = 10000.0f;  g_energy_costs[OP_PROC_SPAWN][2] = 0.0f;
    g_energy_costs[OP_PROC_WAIT][0] = 2.0f;   g_energy_costs[OP_PROC_WAIT][1] = 5000.0f;   g_energy_costs[OP_PROC_WAIT][2] = 0.0f;
    g_energy_costs[OP_PROC_KILL][0] = 1.0f;   g_energy_costs[OP_PROC_KILL][1] = 100.0f;   g_energy_costs[OP_PROC_KILL][2] = 0.0f;
    g_energy_costs[OP_PROC_SIGNAL][0] = 1.0f;  g_energy_costs[OP_PROC_SIGNAL][1] = 50.0f;   g_energy_costs[OP_PROC_SIGNAL][2] = 0.0f;

    /* Utility */
    g_energy_costs[OP_HASH_SHA256][0] = 2.0f;  g_energy_costs[OP_HASH_SHA256][1] = 100.0f;  g_energy_costs[OP_HASH_SHA256][2] = 0.0f;
    g_energy_costs[OP_HASH_MD5][0] = 1.0f;    g_energy_costs[OP_HASH_MD5][1] = 50.0f;     g_energy_costs[OP_HASH_MD5][2] = 0.0f;
    g_energy_costs[OP_COMPRESS_GZIP][0] = 3.0f; g_energy_costs[OP_COMPRESS_GZIP][1] = 500.0f; g_energy_costs[OP_COMPRESS_GZIP][2] = 0.0f;
    g_energy_costs[OP_DECOMPRESS_GZIP][0] = 3.0f; g_energy_costs[OP_DECOMPRESS_GZIP][1] = 500.0f; g_energy_costs[OP_DECOMPRESS_GZIP][2] = 0.0f;
    g_energy_costs[OP_ENCRYPT_AES][0] = 3.0f;  g_energy_costs[OP_ENCRYPT_AES][1] = 200.0f;  g_energy_costs[OP_ENCRYPT_AES][2] = 0.0f;
    g_energy_costs[OP_DECRYPT_AES][0] = 3.0f;  g_energy_costs[OP_DECRYPT_AES][1] = 200.0f;  g_energy_costs[OP_DECRYPT_AES][2] = 0.0f;

    /* Session / Orchestrator */
    g_energy_costs[OP_SESS_BUDGET_CHECK][0] = 1.0f;    g_energy_costs[OP_SESS_BUDGET_CHECK][1] = 1.0f;     g_energy_costs[OP_SESS_BUDGET_CHECK][2] = 0.0f;
    g_energy_costs[OP_SESS_CONTEXT_APPEND][0] = 2.0f;   g_energy_costs[OP_SESS_CONTEXT_APPEND][1] = 5.0f;   g_energy_costs[OP_SESS_CONTEXT_APPEND][2] = 0.0f;
    g_energy_costs[OP_SESS_DENIAL_RECORD][0] = 1.0f;    g_energy_costs[OP_SESS_DENIAL_RECORD][1] = 2.0f;    g_energy_costs[OP_SESS_DENIAL_RECORD][2] = 0.0f;
    g_energy_costs[OP_SESS_SNAPSHOT][0] = 5.0f;         g_energy_costs[OP_SESS_SNAPSHOT][1] = 100.0f;      g_energy_costs[OP_SESS_SNAPSHOT][2] = 0.0f;
    g_energy_costs[OP_SESS_COMPRESS][0] = 3.0f;         g_energy_costs[OP_SESS_COMPRESS][1] = 50.0f;       g_energy_costs[OP_SESS_COMPRESS][2] = 0.0f;
    g_energy_costs[OP_ORCH_CLASSIFY][0] = 2.0f;        g_energy_costs[OP_ORCH_CLASSIFY][1] = 10.0f;       g_energy_costs[OP_ORCH_CLASSIFY][2] = 0.0f;
    g_energy_costs[OP_ORCH_PLAN][0] = 3.0f;             g_energy_costs[OP_ORCH_PLAN][1] = 20.0f;            g_energy_costs[OP_ORCH_PLAN][2] = 0.0f;
    g_energy_costs[OP_ORCH_VALIDATE][0] = 2.0f;         g_energy_costs[OP_ORCH_VALIDATE][1] = 15.0f;         g_energy_costs[OP_ORCH_VALIDATE][2] = 0.0f;
    g_energy_costs[OP_ORCH_EXEC][0] = 5.0f;             g_energy_costs[OP_ORCH_EXEC][1] = 100.0f;           g_energy_costs[OP_ORCH_EXEC][2] = 0.0f;
    g_energy_costs[OP_ORCH_VERIFY][0] = 2.0f;           g_energy_costs[OP_ORCH_VERIFY][1] = 15.0f;           g_energy_costs[OP_ORCH_VERIFY][2] = 0.0f;
    g_energy_costs[OP_ORCH_RESPOND][0] = 2.0f;          g_energy_costs[OP_ORCH_RESPOND][1] = 10.0f;          g_energy_costs[OP_ORCH_RESPOND][2] = 0.0f;

    /* Research */
    g_energy_costs[OP_RESEARCH_HYPOTHESIS_CREATE][0] = 2.0f;     g_energy_costs[OP_RESEARCH_HYPOTHESIS_CREATE][1] = 10.0f;     g_energy_costs[OP_RESEARCH_HYPOTHESIS_CREATE][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_HYPOTHESIS_LOAD][0] = 1.0f;       g_energy_costs[OP_RESEARCH_HYPOTHESIS_LOAD][1] = 5.0f;       g_energy_costs[OP_RESEARCH_HYPOTHESIS_LOAD][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_HYPOTHESIS_INFERENCE][0] = 3.0f;  g_energy_costs[OP_RESEARCH_HYPOTHESIS_INFERENCE][1] = 50.0f;  g_energy_costs[OP_RESEARCH_HYPOTHESIS_INFERENCE][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_EXPERIMENT_RUN][0] = 10.0f;       g_energy_costs[OP_RESEARCH_EXPERIMENT_RUN][1] = 60000.0f;    g_energy_costs[OP_RESEARCH_EXPERIMENT_RUN][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_RESULT_STORE][0] = 3.0f;        g_energy_costs[OP_RESEARCH_RESULT_STORE][1] = 50.0f;        g_energy_costs[OP_RESEARCH_RESULT_STORE][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_STATISTICAL_TEST][0] = 3.0f;      g_energy_costs[OP_RESEARCH_STATISTICAL_TEST][1] = 100.0f;   g_energy_costs[OP_RESEARCH_STATISTICAL_TEST][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_LITERATURE_FETCH][0] = 5.0f;     g_energy_costs[OP_RESEARCH_LITERATURE_FETCH][1] = 10000.0f; g_energy_costs[OP_RESEARCH_LITERATURE_FETCH][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_LITERATURE_PARSE][0] = 5.0f;    g_energy_costs[OP_RESEARCH_LITERATURE_PARSE][1] = 500.0f;   g_energy_costs[OP_RESEARCH_LITERATURE_PARSE][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_LITERATURE_INDEX][0] = 5.0f;    g_energy_costs[OP_RESEARCH_LITERATURE_INDEX][1] = 200.0f;   g_energy_costs[OP_RESEARCH_LITERATURE_INDEX][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_CITATION_LINK][0] = 2.0f;         g_energy_costs[OP_RESEARCH_CITATION_LINK][1] = 20.0f;       g_energy_costs[OP_RESEARCH_CITATION_LINK][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_LITERATURE_EMBED][0] = 10.0f;    g_energy_costs[OP_RESEARCH_LITERATURE_EMBED][1] = 5000.0f;   g_energy_costs[OP_RESEARCH_LITERATURE_EMBED][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_PROGRESS_STORE][0] = 2.0f;      g_energy_costs[OP_RESEARCH_PROGRESS_STORE][1] = 30.0f;      g_energy_costs[OP_RESEARCH_PROGRESS_STORE][2] = 0.0f;
    g_energy_costs[OP_RESEARCH_CONTEXT_SUMMARIZE][0] = 5.0f;   g_energy_costs[OP_RESEARCH_CONTEXT_SUMMARIZE][1] = 200.0f;   g_energy_costs[OP_RESEARCH_CONTEXT_SUMMARIZE][2] = 0.0f;

    /* Self-Management */
    g_energy_costs[OP_SELF_CHECKPOINT_CREATE][0] = 5.0f;   g_energy_costs[OP_SELF_CHECKPOINT_CREATE][1] = 100.0f;   g_energy_costs[OP_SELF_CHECKPOINT_CREATE][2] = 0.0f;
    g_energy_costs[OP_SELF_CHECKPOINT_RESTORE][0] = 3.0f;  g_energy_costs[OP_SELF_CHECKPOINT_RESTORE][1] = 50.0f;    g_energy_costs[OP_SELF_CHECKPOINT_RESTORE][2] = 0.0f;
    g_energy_costs[OP_SELF_BUDGET_REALLOCATE][0] = 1.0f;   g_energy_costs[OP_SELF_BUDGET_REALLOCATE][1] = 5.0f;      g_energy_costs[OP_SELF_BUDGET_REALLOCATE][2] = 0.0f;
    g_energy_costs[OP_SELF_STRATEGY_PIVOT][0] = 2.0f;      g_energy_costs[OP_SELF_STRATEGY_PIVOT][1] = 10.0f;       g_energy_costs[OP_SELF_STRATEGY_PIVOT][2] = 0.0f;
    g_energy_costs[OP_SELF_PROGRESS_ASSESS][0] = 1.0f;     g_energy_costs[OP_SELF_PROGRESS_ASSESS][1] = 5.0f;       g_energy_costs[OP_SELF_PROGRESS_ASSESS][2] = 0.0f;
    g_energy_costs[OP_SELF_CONTEXT_SUMMARIZE][0] = 5.0f;   g_energy_costs[OP_SELF_CONTEXT_SUMMARIZE][1] = 200.0f;   g_energy_costs[OP_SELF_CONTEXT_SUMMARIZE][2] = 0.0f;
}

static const char* arg_value_string(OpPacketEx* packet, const char* key) {
    for (uint32_t i = 0; i < packet->arg_count; i++) {
        if (packet->args[i].type == ARG_TYPE_STRING && strcmp(packet->args[i].key, key) == 0) {
            return packet->args[i].value.s;
        }
    }
    return NULL;
}

static int64_t arg_value_int(OpPacketEx* packet, const char* key, int64_t default_val) {
    for (uint32_t i = 0; i < packet->arg_count; i++) {
        if (packet->args[i].type == ARG_TYPE_INT && strcmp(packet->args[i].key, key) == 0) {
            return packet->args[i].value.i;
        }
    }
    return default_val;
}

static bool arg_value_bool(OpPacketEx* packet, const char* key, bool default_val) {
    for (uint32_t i = 0; i < packet->arg_count; i++) {
        if (packet->args[i].type == ARG_TYPE_BOOL && strcmp(packet->args[i].key, key) == 0) {
            return packet->args[i].value.b;
        }
    }
    return default_val;
}

/* ============================================================================
 * Permission stubs (session layer integration points)
 * ============================================================================ */
bool session_has_explicit_allow(OpCode opcode, uint32_t chain_id) {
    (void)opcode;
    (void)chain_id;
    return false; /* Default deny -- session layer overrides */
}

bool session_has_2vote_verify(uint32_t chain_id) {
    (void)chain_id;
    return false; /* Default deny -- session layer overrides */
}

/* ============================================================================
 * Lifecycle
 * ============================================================================ */
int ops_init(void) {
    if (g_initialized) {
        return ERR_NOT_INIT;
    }
    memset(g_op_registry, 0, sizeof(g_op_registry));
    init_conflict_matrix();
    init_energy_costs();
    g_initialized = true;
    return ERR_OK;
}

void ops_shutdown(void) {
    if (!g_initialized) {
        return;
    }
    memset(g_op_registry, 0, sizeof(g_op_registry));
    memset(g_conflict_matrix, 0, sizeof(g_conflict_matrix));
    memset(g_energy_costs, 0, sizeof(g_energy_costs));
    g_initialized = false;
}

int ops_register(OpCodeDef* def) {
    if (!g_initialized || !def || def->opcode >= OP_MAX) {
        return ERR_BAD_ARGS;
    }
    g_op_registry[def->opcode] = *def;
    return ERR_OK;
}

const OpCodeDef* ops_get_definition(OpCode opcode) {
    if (!g_initialized || opcode >= OP_MAX) {
        return NULL;
    }
    return &g_op_registry[opcode];
}

/* ============================================================================
 * String / opcode mapping
 * ============================================================================ */
static const struct { OpCode code; const char* name; } g_opcode_names[] = {
    { OP_NOP, "NOP" },
    { OP_MMAP_ALLOC, "MMAP_ALLOC" }, { OP_MMAP_FREE, "MMAP_FREE" },
    { OP_MMAP_READ, "MMAP_READ" }, { OP_MMAP_WRITE, "MMAP_WRITE" },
    { OP_MMAP_SYNC, "MMAP_SYNC" },
    { OP_IO_READ, "IO_READ" }, { OP_IO_WRITE, "IO_WRITE" },
    { OP_IO_OPEN, "IO_OPEN" }, { OP_IO_CLOSE, "IO_CLOSE" },
    { OP_IO_SEEK, "IO_SEEK" },
    { OP_GIT_INIT, "GIT_INIT" }, { OP_GIT_CLONE, "GIT_CLONE" },
    { OP_GIT_FETCH, "GIT_FETCH" }, { OP_GIT_STATUS, "GIT_STATUS" },
    { OP_GIT_DIFF, "GIT_DIFF" }, { OP_GIT_ADD, "GIT_ADD" },
    { OP_GIT_COMMIT, "GIT_COMMIT" }, { OP_GIT_PUSH, "GIT_PUSH" },
    { OP_GIT_CHECKOUT, "GIT_CHECKOUT" }, { OP_GIT_BRANCH, "GIT_BRANCH" },
    { OP_GIT_MERGE, "GIT_MERGE" }, { OP_GIT_REBASE, "GIT_REBASE" },
    { OP_GIT_TAG, "GIT_TAG" }, { OP_GIT_RESET, "GIT_RESET" },
    { OP_BUILD_COMPILE, "BUILD_COMPILE" }, { OP_BUILD_LINK, "BUILD_LINK" },
    { OP_BUILD_TEST, "BUILD_TEST" }, { OP_BUILD_DEPLOY, "BUILD_DEPLOY" },
    { OP_BUILD_CLEAN, "BUILD_CLEAN" },
    { OP_NET_HTTP_GET, "NET_HTTP_GET" }, { OP_NET_HTTP_POST, "NET_HTTP_POST" },
    { OP_NET_TCP_CONNECT, "NET_TCP_CONNECT" }, { OP_NET_TCP_SEND, "NET_TCP_SEND" },
    { OP_NET_TCP_RECV, "NET_TCP_RECV" }, { OP_NET_TCP_CLOSE, "NET_TCP_CLOSE" },
    { OP_NET_WEBSOCKET, "NET_WEBSOCKET" },
    { OP_PROC_SPAWN, "PROC_SPAWN" }, { OP_PROC_WAIT, "PROC_WAIT" },
    { OP_PROC_KILL, "PROC_KILL" }, { OP_PROC_SIGNAL, "PROC_SIGNAL" },
    { OP_HASH_SHA256, "HASH_SHA256" }, { OP_HASH_MD5, "HASH_MD5" },
    { OP_COMPRESS_GZIP, "COMPRESS_GZIP" }, { OP_DECOMPRESS_GZIP, "DECOMPRESS_GZIP" },
    { OP_ENCRYPT_AES, "ENCRYPT_AES" }, { OP_DECRYPT_AES, "DECRYPT_AES" },
    { OP_SYS_EXEC, "SYS_EXEC" }, { OP_SYS_ENV_GET, "SYS_ENV_GET" },
    { OP_SYS_ENV_SET, "SYS_ENV_SET" }, { OP_SYS_FILE_EXISTS, "SYS_FILE_EXISTS" },
    { OP_SYS_DIR_CREATE, "SYS_DIR_CREATE" }, { OP_SYS_DIR_REMOVE, "SYS_DIR_REMOVE" },
    { OP_SYS_FILE_COPY, "SYS_FILE_COPY" }, { OP_SYS_FILE_MOVE, "SYS_FILE_MOVE" },
    { OP_SYS_FILE_DELETE, "SYS_FILE_DELETE" }, { OP_SYS_CHMOD, "SYS_CHMOD" },
    { OP_SESS_BUDGET_CHECK, "SESS_BUDGET_CHECK" },
    { OP_SESS_CONTEXT_APPEND, "SESS_CONTEXT_APPEND" },
    { OP_SESS_DENIAL_RECORD, "SESS_DENIAL_RECORD" },
    { OP_SESS_SNAPSHOT, "SESS_SNAPSHOT" },
    { OP_SESS_COMPRESS, "SESS_COMPRESS" },
    { OP_ORCH_CLASSIFY, "ORCH_CLASSIFY" },
    { OP_ORCH_PLAN, "ORCH_PLAN" },
    { OP_ORCH_VALIDATE, "ORCH_VALIDATE" },
    { OP_ORCH_EXEC, "ORCH_EXEC" },
    { OP_ORCH_VERIFY, "ORCH_VERIFY" },
    { OP_ORCH_RESPOND, "ORCH_RESPOND" },
    { OP_RESEARCH_HYPOTHESIS_CREATE, "RESEARCH_HYPOTHESIS_CREATE" },
    { OP_RESEARCH_HYPOTHESIS_LOAD, "RESEARCH_HYPOTHESIS_LOAD" },
    { OP_RESEARCH_HYPOTHESIS_INFERENCE, "RESEARCH_HYPOTHESIS_INFERENCE" },
    { OP_RESEARCH_EXPERIMENT_RUN, "RESEARCH_EXPERIMENT_RUN" },
    { OP_RESEARCH_RESULT_STORE, "RESEARCH_RESULT_STORE" },
    { OP_RESEARCH_STATISTICAL_TEST, "RESEARCH_STATISTICAL_TEST" },
    { OP_RESEARCH_LITERATURE_FETCH, "RESEARCH_LITERATURE_FETCH" },
    { OP_RESEARCH_LITERATURE_PARSE, "RESEARCH_LITERATURE_PARSE" },
    { OP_RESEARCH_LITERATURE_INDEX, "RESEARCH_LITERATURE_INDEX" },
    { OP_RESEARCH_CITATION_LINK, "RESEARCH_CITATION_LINK" },
    { OP_RESEARCH_LITERATURE_EMBED, "RESEARCH_LITERATURE_EMBED" },
    { OP_RESEARCH_PROGRESS_STORE, "RESEARCH_PROGRESS_STORE" },
    { OP_RESEARCH_CONTEXT_SUMMARIZE, "RESEARCH_CONTEXT_SUMMARIZE" },
    { OP_SELF_CHECKPOINT_CREATE, "SELF_CHECKPOINT_CREATE" },
    { OP_SELF_CHECKPOINT_RESTORE, "SELF_CHECKPOINT_RESTORE" },
    { OP_SELF_BUDGET_REALLOCATE, "SELF_BUDGET_REALLOCATE" },
    { OP_SELF_STRATEGY_PIVOT, "SELF_STRATEGY_PIVOT" },
    { OP_SELF_PROGRESS_ASSESS, "SELF_PROGRESS_ASSESS" },
    { OP_SELF_CONTEXT_SUMMARIZE, "SELF_CONTEXT_SUMMARIZE" },
    { OP_MAX, NULL }
};

const char* ops_opcode_to_string(OpCode opcode) {
    for (size_t i = 0; g_opcode_names[i].name != NULL; i++) {
        if (g_opcode_names[i].code == opcode) {
            return g_opcode_names[i].name;
        }
    }
    return "UNKNOWN";
}

OpCode ops_string_to_opcode(const char* name) {
    if (!name) return OP_NOP;
    for (size_t i = 0; g_opcode_names[i].name != NULL; i++) {
        if (strcmp(g_opcode_names[i].name, name) == 0) {
            return g_opcode_names[i].code;
        }
    }
    return OP_NOP;
}

/* ============================================================================
 * Packet helpers
 * ============================================================================ */
void ops_packet_init(OpPacketEx* packet, OpCode opcode) {
    if (!packet) return;
    memset(packet, 0, sizeof(OpPacketEx));
    packet->opcode = (uint8_t)opcode;
    packet->fd_in = -1;
    packet->fd_out = -1;
}

void ops_packet_set_string(OpPacketEx* packet, const char* key, const char* value) {
    if (!packet || !key || !value || packet->arg_count >= MAX_ARGS) return;
    OpArg* arg = &packet->args[packet->arg_count++];
    strncpy(arg->key, key, sizeof(arg->key) - 1);
    arg->type = ARG_TYPE_STRING;
    strncpy(arg->value.s, value, sizeof(arg->value.s) - 1);
}

void ops_packet_set_int(OpPacketEx* packet, const char* key, int64_t value) {
    if (!packet || !key || packet->arg_count >= MAX_ARGS) return;
    OpArg* arg = &packet->args[packet->arg_count++];
    strncpy(arg->key, key, sizeof(arg->key) - 1);
    arg->type = ARG_TYPE_INT;
    arg->value.i = value;
}

/* ============================================================================
 * Memory operations
 * ============================================================================ */
void* ops_mmap_alloc(size_t size) {
    if (size == 0) return NULL;
    void* ptr = mmap(NULL, size, PROT_READ | PROT_WRITE, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    if (ptr == MAP_FAILED) return NULL;
    return ptr;
}

int ops_mmap_free(void* ptr, size_t size) {
    if (!ptr || size == 0) return ERR_BAD_ARGS;
    if (munmap(ptr, size) != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

int ops_mmap_sync(void* ptr, size_t size) {
    if (!ptr || size == 0) return ERR_BAD_ARGS;
    if (msync(ptr, size, MS_SYNC) != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

/* ============================================================================
 * Time
 * ============================================================================ */
uint64_t ops_get_time_ns(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

/* ============================================================================
 * State hash
 * ============================================================================ */
uint64_t ops_compute_state_hash(ExecContext* ctx) {
    if (!ctx) return 0;
    uint64_t h = 14695981039346656037ULL;
    for (size_t i = 0; i < ctx->fd_count; i++) {
        h ^= (uint64_t)ctx->open_fds[i];
        h *= 1099511628211ULL;
    }
    for (size_t i = 0; i < ctx->mmap_count; i++) {
        h ^= (uint64_t)(uintptr_t)ctx->mmap_regions[i];
        h *= 1099511628211ULL;
        h ^= (uint64_t)ctx->mmap_sizes[i];
        h *= 1099511628211ULL;
    }
    return h;
}

/* ============================================================================
 * Energy calculation
 * ============================================================================ */
float ops_calculate_action(const OpPacketEx* packets, size_t count) {
    if (!packets) return 0.0f;
    float total = 0.0f;
    for (size_t i = 0; i < count; i++) {
        total += g_energy_costs[packets[i].opcode][0];
    }
    return total;
}

/* ============================================================================
 * Conflict check
 * ============================================================================ */
bool ops_check_conflict(OpCode a, OpCode b) {
    if (a >= 256 || b >= 256) return true;
    return g_conflict_matrix[a][b] > CONFLICT_NONE;
}

/* ============================================================================
 * Validation
 * ============================================================================ */
ValidationResult ops_validate_chain(const OpPacketEx* packets, size_t count, ExecContext* ctx) {
    ValidationResult result = {0};
    result.is_valid = false;
    result.error_code = ERR_OK;
    result.invalid_op_index = 0;
    result.conflict_op_pair[0] = 0;
    result.conflict_op_pair[1] = 0;

    /* Step 1: State check */
    if (!g_initialized) {
        result.error_code = ERR_NOT_INIT;
        snprintf(result.error_msg, ERR_MSG_LEN, "Core not initialized");
        return result;
    }
    if (count == 0) {
        result.error_code = ERR_INVALID_CHAIN;
        snprintf(result.error_msg, ERR_MSG_LEN, "Empty chain");
        return result;
    }
    if (!packets) {
        result.error_code = ERR_INVALID_CHAIN;
        snprintf(result.error_msg, ERR_MSG_LEN, "Null packet pointer");
        return result;
    }

    /* Step 2: Opcode validity */
    for (size_t i = 0; i < count; i++) {
        if (i > 0 && packets[i].opcode == OP_NOP) {
            result.error_code = ERR_INVALID_OPCODE;
            result.invalid_op_index = (uint32_t)i;
            snprintf(result.error_msg, ERR_MSG_LEN, "NOP at index %zu", i);
            return result;
        }
        if (packets[i].opcode >= OP_MAX) {
            result.error_code = ERR_INVALID_OPCODE;
            result.invalid_op_index = (uint32_t)i;
            snprintf(result.error_msg, ERR_MSG_LEN, "Invalid opcode %u at index %zu", packets[i].opcode, i);
            return result;
        }
    }

    /* Step 3: Registration check */
    for (size_t i = 0; i < count; i++) {
        if (g_op_registry[packets[i].opcode].execute == NULL) {
            result.error_code = ERR_UNREGISTERED;
            result.invalid_op_index = (uint32_t)i;
            snprintf(result.error_msg, ERR_MSG_LEN, "Opcode %s not registered", ops_opcode_to_string(packets[i].opcode));
            return result;
        }
    }

    /* Step 4: Argument validity */
    for (size_t i = 0; i < count; i++) {
        if (packets[i].arg_count > MAX_ARGS) {
            result.error_code = ERR_BAD_ARGS;
            result.invalid_op_index = (uint32_t)i;
            snprintf(result.error_msg, ERR_MSG_LEN, "arg_count %u > MAX_ARGS", packets[i].arg_count);
            return result;
        }
        for (uint32_t j = 0; j < packets[i].arg_count; j++) {
            if (packets[i].args[j].key[0] == '\0') {
                result.error_code = ERR_BAD_ARGS;
                result.invalid_op_index = (uint32_t)i;
                snprintf(result.error_msg, ERR_MSG_LEN, "Empty key at arg %u", j);
                return result;
            }
            if (packets[i].args[j].type > ARG_TYPE_BLOB) {
                result.error_code = ERR_BAD_ARGS;
                result.invalid_op_index = (uint32_t)i;
                snprintf(result.error_msg, ERR_MSG_LEN, "Invalid type %u at arg %u", packets[i].args[j].type, j);
                return result;
            }
            if (packets[i].args[j].type == ARG_TYPE_BLOB) {
                if (packets[i].args[j].value.blob.buffer == NULL || packets[i].args[j].value.blob.buffer_size == 0) {
                    result.error_code = ERR_BAD_ARGS;
                    result.invalid_op_index = (uint32_t)i;
                    snprintf(result.error_msg, ERR_MSG_LEN, "Blob without buffer at arg %u", j);
                    return result;
                }
            }
        }
        /* Duplicate key check */
        for (uint32_t j = 0; j < packets[i].arg_count; j++) {
            for (uint32_t k = j + 1; k < packets[i].arg_count; k++) {
                if (strcmp(packets[i].args[j].key, packets[i].args[k].key) == 0) {
                    result.error_code = ERR_BAD_ARGS;
                    result.invalid_op_index = (uint32_t)i;
                    snprintf(result.error_msg, ERR_MSG_LEN, "Duplicate key '%s'", packets[i].args[j].key);
                    return result;
                }
            }
        }
    }

    /* Step 5: FD validity */
    for (size_t i = 0; i < count; i++) {
        if (packets[i].fd_in < -1 || packets[i].fd_out < -1) {
            result.error_code = ERR_BAD_FD;
            result.invalid_op_index = (uint32_t)i;
            snprintf(result.error_msg, ERR_MSG_LEN, "Negative FD (other than -1)");
            return result;
        }
        if (packets[i].fd_in == packets[i].fd_out && packets[i].fd_in != -1) {
            result.error_code = ERR_BAD_FD;
            result.invalid_op_index = (uint32_t)i;
            snprintf(result.error_msg, ERR_MSG_LEN, "fd_in == fd_out");
            return result;
        }
        if (packets[i].opcode == OP_IO_OPEN && packets[i].fd_in != -1) {
            result.error_code = ERR_BAD_FD;
            result.invalid_op_index = (uint32_t)i;
            snprintf(result.error_msg, ERR_MSG_LEN, "OP_IO_OPEN with fd_in != -1");
            return result;
        }
    }

    /* Step 6: Pairwise conflict check */
    for (size_t i = 0; i < count; i++) {
        for (size_t j = i + 1; j < count; j++) {
            uint8_t level = g_conflict_matrix[packets[i].opcode][packets[j].opcode];
            if (level >= CONFLICT_MEDIUM) {
                result.error_code = ERR_CONFLICT;
                result.conflict_op_pair[0] = (uint32_t)i;
                result.conflict_op_pair[1] = (uint32_t)j;
                snprintf(result.error_msg, ERR_MSG_LEN,
                         "Conflict between op %zu (%s) and op %zu (%s): level %u",
                         i, ops_opcode_to_string(packets[i].opcode),
                         j, ops_opcode_to_string(packets[j].opcode),
                         level);
                return result;
            }
        }
    }

    /* Step 7: Energy budget check */
    float total_energy = 0.0f;
    float total_latency_us = 0.0f;
    for (size_t i = 0; i < count; i++) {
        total_energy += g_energy_costs[packets[i].opcode][0];
        total_latency_us += g_energy_costs[packets[i].opcode][1];
    }
    result.total_energy = total_energy;
    result.estimated_latency_us = total_latency_us;

    if (ctx) {
        if (total_energy > ctx->session_budget_tokens) {
            result.error_code = ERR_ENERGY_OVERFLOW;
            snprintf(result.error_msg, ERR_MSG_LEN, "Energy %.2f > budget %.2f", total_energy, ctx->session_budget_tokens);
            return result;
        }
        if (total_latency_us > ctx->session_budget_time_ms * 1000.0f) {
            result.error_code = ERR_ENERGY_OVERFLOW;
            snprintf(result.error_msg, ERR_MSG_LEN, "Latency %.2f us > budget %.2f us", total_latency_us, ctx->session_budget_time_ms * 1000.0f);
            return result;
        }
    }

    /* Step 8: Permission check */
    for (size_t i = 0; i < count; i++) {
        const OpCodeDef* def = &g_op_registry[packets[i].opcode];
        if ((def->required_flags & OP_FLAG_DANGEROUS) || (packets[i].flags & OP_FLAG_DANGEROUS)) {
            if (!session_has_explicit_allow(packets[i].opcode, ctx ? ctx->chain_id : 0)) {
                result.error_code = ERR_PERMISSION;
                result.invalid_op_index = (uint32_t)i;
                snprintf(result.error_msg, ERR_MSG_LEN, "Permission denied: dangerous op %s", def->name);
                return result;
            }
        }
        if (def->safety_level == 0) {
            if (!session_has_2vote_verify(ctx ? ctx->chain_id : 0)) {
                result.error_code = ERR_PERMISSION;
                result.invalid_op_index = (uint32_t)i;
                snprintf(result.error_msg, ERR_MSG_LEN, "Permission denied: critical op %s requires 2-vote", def->name);
                return result;
            }
        }
    }
    if (ctx && ctx->circuit_broken) {
        result.error_code = ERR_CIRCUIT_BROKEN;
        snprintf(result.error_msg, ERR_MSG_LEN, "Circuit broken: manual reset required");
        return result;
    }

    /* All checks passed */
    result.is_valid = true;
    result.error_code = ERR_OK;
    return result;
}

/* ============================================================================
 * Backup & cleanup
 * ============================================================================ */
int ops_create_backup(const char* path, char* backup_path, size_t backup_path_size) {
    if (!path || !backup_path || backup_path_size == 0) return ERR_BAD_ARGS;
    struct stat st;
    if (stat(path, &st) != 0) return ERR_OK; /* nothing to backup */

    uint64_t ts = ops_get_time_ns();
    snprintf(backup_path, backup_path_size, ".mimic/backups/%s.%lu.bak", path, (unsigned long)ts);

    /* Ensure directory exists */
    char dir[512];
    strncpy(dir, backup_path, sizeof(dir) - 1);
    dir[sizeof(dir) - 1] = '\0';
    char* d = dirname(dir);
    (void)mkdir(d, 0755);

    FILE* src = fopen(path, "rb");
    if (!src) return ERR_EXEC_FAIL;
    FILE* dst = fopen(backup_path, "wb");
    if (!dst) { fclose(src); return ERR_EXEC_FAIL; }

    char buf[8192];
    size_t n;
    while ((n = fread(buf, 1, sizeof(buf), src)) > 0) {
        if (fwrite(buf, 1, n, dst) != n) {
            fclose(src); fclose(dst);
            return ERR_EXEC_FAIL;
        }
    }
    fclose(src);
    fclose(dst);
    return ERR_OK;
}

void ops_best_effort_cleanup(OpPacketEx* packet, ExecContext* ctx) {
    (void)ctx;
    if (!packet) return;
    switch (packet->opcode) {
        case OP_IO_WRITE:
        case OP_SYS_FILE_DELETE:
        case OP_SYS_DIR_REMOVE:
        case OP_SYS_FILE_MOVE:
        case OP_BUILD_COMPILE:
        case OP_BUILD_CLEAN:
        case OP_GIT_COMMIT:
        case OP_GIT_BRANCH:
        case OP_GIT_CHECKOUT:
        case OP_PROC_SPAWN:
        case OP_NET_HTTP_POST:
            /* Best-effort: log irreversible action. Real cleanup depends on backup. */
            break;
        case OP_SYS_FILE_COPY:
            /* Delete the copied file */
            {
                const char* dst = arg_value_string(packet, "dst");
                if (dst) unlink(dst);
            }
            break;
        default:
            break;
    }
}

/* ============================================================================
 * Rollback
 * ============================================================================ */
int ops_rollback_chain(OpPacketEx* packets, uint32_t failed_index, ExecContext* ctx) {
    if (!ctx || !packets) return ERR_ROLLBACK_FAIL;
    if (!ctx->pre_state_blob) {
        return ERR_ROLLBACK_FAIL;
    }

    /* Phase 1: Execute inverses for reversible ops */
    for (int j = (int)failed_index - 1; j >= 0; j--) {
        OpCodeDef* def = &g_op_registry[packets[j].opcode];
        if (def->is_reversible && def->inverse_execute) {
            OpPacketEx inverse;
            ops_packet_init(&inverse, def->inverse_opcode);
            int inv_result = def->inverse_execute(&packets[j], &inverse);
            if (inv_result != ERR_OK) {
                /* Inverse failed = partial rollback. Continue with best-effort. */
            }
        } else {
            ops_best_effort_cleanup(&packets[j], ctx);
        }
    }

    /* Phase 2: Resource cleanup */
    for (size_t k = 0; k < ctx->fd_count; k++) {
        close(ctx->open_fds[k]);
    }
    ctx->fd_count = 0;

    for (size_t k = 0; k < ctx->mmap_count; k++) {
        ops_mmap_sync(ctx->mmap_regions[k], ctx->mmap_sizes[k]);
        ops_mmap_free(ctx->mmap_regions[k], ctx->mmap_sizes[k]);
    }
    ctx->mmap_count = 0;

    /* Phase 3: State verification */
    uint64_t current_hash = ops_compute_state_hash(ctx);
    if (current_hash == ctx->pre_state_hash) {
        return ERR_OK;
    } else {
        return ERR_ROLLBACK_FAIL;
    }
}

/* ============================================================================
 * Domain executors
 * ============================================================================ */

/* --- NOP --- */
static int exec_nop(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

/* --- Memory --- */
static int exec_mmap_alloc(OpPacketEx* packet) {
    int64_t size = arg_value_int(packet, "size", 0);
    if (size <= 0) return ERR_BAD_ARGS;
    void* ptr = ops_mmap_alloc((size_t)size);
    if (!ptr) return ERR_EXEC_FAIL;
    /* Return pointer via result slot (simplified) */
    return ERR_OK;
}

static int exec_mmap_free(OpPacketEx* packet) {
    int64_t ptr_val = arg_value_int(packet, "ptr", 0);
    int64_t size = arg_value_int(packet, "size", 0);
    if (ptr_val == 0 || size <= 0) return ERR_BAD_ARGS;
    if (munmap((void*)(uintptr_t)ptr_val, (size_t)size) != 0) {
        return ERR_EXEC_FAIL;
    }
    return ERR_OK;
}

static int exec_mmap_read(OpPacketEx* packet) {
    int64_t ptr_val = arg_value_int(packet, "ptr", 0);
    int64_t offset = arg_value_int(packet, "offset", 0);
    int64_t length = arg_value_int(packet, "length", 0);
    if (ptr_val == 0 || length <= 0 || length > 4096) return ERR_BAD_ARGS;
    char* src = (char*)((uintptr_t)ptr_val + (size_t)offset);
    size_t to_copy = (size_t)length;
    if (to_copy > sizeof(packet->result) - 1) to_copy = sizeof(packet->result) - 1;
    memcpy(packet->result, src, to_copy);
    packet->result[to_copy] = 0;
    packet->result_len = to_copy;
    return ERR_OK;
}

static int exec_mmap_write(OpPacketEx* packet) {
    int64_t ptr_val = arg_value_int(packet, "ptr", 0);
    int64_t offset = arg_value_int(packet, "offset", 0);
    const char* data = arg_value_string(packet, "data");
    if (ptr_val == 0 || !data) return ERR_BAD_ARGS;
    size_t len = strlen(data);
    if (len > 4096) len = 4096;
    char* dst = (char*)((uintptr_t)ptr_val + (size_t)offset);
    memcpy(dst, data, len);
    return ERR_OK;
}

static int exec_mmap_sync(OpPacketEx* packet) {
    int64_t ptr_val = arg_value_int(packet, "ptr", 0);
    int64_t size = arg_value_int(packet, "size", 0);
    if (ptr_val == 0 || size <= 0) return ERR_BAD_ARGS;
    if (msync((void*)(uintptr_t)ptr_val, (size_t)size, MS_SYNC) != 0) {
        return ERR_EXEC_FAIL;
    }
    return ERR_OK;
}

/* --- I/O --- */
static int exec_io_open(OpPacketEx* packet) {
    const char* path = arg_value_string(packet, "path");
    const char* mode = arg_value_string(packet, "mode");
    if (!path || !mode) return ERR_BAD_ARGS;

    int flags = 0;
    if (strcmp(mode, "r") == 0 || strcmp(mode, "rb") == 0) {
        flags = O_RDONLY;
    } else if (strcmp(mode, "w") == 0 || strcmp(mode, "wb") == 0) {
        flags = O_WRONLY | O_CREAT | O_TRUNC;
    } else if (strcmp(mode, "a") == 0 || strcmp(mode, "ab") == 0) {
        flags = O_WRONLY | O_CREAT | O_APPEND;
    } else if (strcmp(mode, "rw") == 0 || strcmp(mode, "r+") == 0) {
        flags = O_RDWR;
    } else {
        return ERR_BAD_ARGS;
    }

    int fd = open(path, flags, 0644);
    if (fd < 0) return ERR_EXEC_FAIL;
    packet->fd_out = fd;
    return ERR_OK;
}

static int exec_io_close(OpPacketEx* packet) {
    int fd = packet->fd_in >= 0 ? packet->fd_in : arg_value_int(packet, "fd", -1);
    if (fd < 0) return ERR_BAD_ARGS;
    if (close(fd) != 0) return ERR_EXEC_FAIL;
    packet->fd_in = -1;
    return ERR_OK;
}

static int exec_io_read(OpPacketEx* packet) {
    int fd = packet->fd_in >= 0 ? packet->fd_in : arg_value_int(packet, "fd", -1);
    int64_t length = arg_value_int(packet, "length", 0);
    if (fd < 0 || length <= 0) return ERR_BAD_ARGS;
    if (length > (int64_t)sizeof(packet->result) - 1) length = (int64_t)sizeof(packet->result) - 1;
    ssize_t n = read(fd, packet->result, (size_t)length);
    if (n < 0) return ERR_EXEC_FAIL;
    packet->result[n] = 0;
    packet->result_len = (size_t)n;
    return ERR_OK;
}

static int exec_io_write(OpPacketEx* packet) {
    int fd = packet->fd_in >= 0 ? packet->fd_in : arg_value_int(packet, "fd", -1);
    if (fd < 0) return ERR_BAD_ARGS;
    const char* data = arg_value_string(packet, "data");
    if (!data) return ERR_BAD_ARGS;
    size_t len = strlen(data);
    ssize_t written = write(fd, data, len);
    if ((size_t)written != len) return ERR_EXEC_FAIL;
    return ERR_OK;
}

static int exec_io_seek(OpPacketEx* packet) {
    int fd = packet->fd_in >= 0 ? packet->fd_in : arg_value_int(packet, "fd", -1);
    int64_t offset = arg_value_int(packet, "offset", 0);
    int64_t whence = arg_value_int(packet, "whence", 0);
    if (fd < 0) return ERR_BAD_ARGS;
    int w = (whence == 0) ? SEEK_SET : (whence == 1) ? SEEK_CUR : SEEK_END;
    if (lseek(fd, (off_t)offset, w) < 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

/* --- System --- */
static int exec_sys_file_exists(OpPacketEx* packet) {
    const char* path = arg_value_string(packet, "path");
    if (!path) return ERR_BAD_ARGS;
    struct stat st;
    bool exists = (stat(path, &st) == 0);
    /* Result would be returned via output mechanism. Simplified. */
    (void)exists;
    return ERR_OK;
}

static int exec_sys_dir_create(OpPacketEx* packet) {
    const char* path = arg_value_string(packet, "path");
    if (!path) return ERR_BAD_ARGS;
    int64_t mode = arg_value_int(packet, "mode", 0755);
    bool recursive = arg_value_bool(packet, "recursive", false);
    if (recursive) {
        /* Simple recursive mkdir simulation */
        char tmp[512];
        strncpy(tmp, path, sizeof(tmp) - 1);
        tmp[sizeof(tmp) - 1] = '\0';
        for (char* p = tmp + 1; *p; p++) {
            if (*p == '/') {
                *p = '\0';
                (void)mkdir(tmp, (mode_t)mode);
                *p = '/';
            }
        }
        (void)mkdir(tmp, (mode_t)mode);
    } else {
        if (mkdir(path, (mode_t)mode) != 0 && errno != EEXIST) {
            return ERR_EXEC_FAIL;
        }
    }
    return ERR_OK;
}

static int exec_sys_dir_remove(OpPacketEx* packet) {
    const char* path = arg_value_string(packet, "path");
    if (!path) return ERR_BAD_ARGS;
    bool recursive = arg_value_bool(packet, "recursive", false);
    if (recursive) {
        /* Simplified: rmdir only supports empty dirs in POSIX. Real recursive needs tree walk. */
        return ERR_EXEC_FAIL; /* Not fully implemented */
    } else {
        if (rmdir(path) != 0) return ERR_EXEC_FAIL;
    }
    return ERR_OK;
}

static int exec_sys_file_copy(OpPacketEx* packet) {
    const char* src = arg_value_string(packet, "src");
    const char* dst = arg_value_string(packet, "dst");
    if (!src || !dst) return ERR_BAD_ARGS;

    FILE* s = fopen(src, "rb");
    if (!s) return ERR_EXEC_FAIL;
    FILE* d = fopen(dst, "wb");
    if (!d) { fclose(s); return ERR_EXEC_FAIL; }

    char buf[8192];
    size_t n;
    while ((n = fread(buf, 1, sizeof(buf), s)) > 0) {
        if (fwrite(buf, 1, n, d) != n) {
            fclose(s); fclose(d);
            return ERR_EXEC_FAIL;
        }
    }
    fclose(s);
    fclose(d);
    return ERR_OK;
}

static int exec_sys_file_move(OpPacketEx* packet) {
    const char* src = arg_value_string(packet, "src");
    const char* dst = arg_value_string(packet, "dst");
    if (!src || !dst) return ERR_BAD_ARGS;
    if (rename(src, dst) != 0) {
        /* Fallback to copy+delete */
        int r = exec_sys_file_copy(packet);
        if (r != ERR_OK) return r;
        if (unlink(src) != 0) return ERR_EXEC_FAIL;
    }
    return ERR_OK;
}

static int exec_sys_file_delete(OpPacketEx* packet) {
    const char* path = arg_value_string(packet, "path");
    if (!path) return ERR_BAD_ARGS;
    if (unlink(path) != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

static int exec_sys_chmod(OpPacketEx* packet) {
    const char* path = arg_value_string(packet, "path");
    int64_t mode = arg_value_int(packet, "mode", 0);
    if (!path) return ERR_BAD_ARGS;
    /* Safety: block setuid/setgid */
    if (mode & (S_ISUID | S_ISGID)) {
        return ERR_PERMISSION;
    }
    if (chmod(path, (mode_t)mode) != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

static int exec_sys_env_get(OpPacketEx* packet) {
    const char* name = arg_value_string(packet, "name");
    if (!name) return ERR_BAD_ARGS;
    const char* val = getenv(name);
    (void)val; /* Would set result */
    return ERR_OK;
}

static int exec_sys_env_set(OpPacketEx* packet) {
    const char* name = arg_value_string(packet, "name");
    const char* value = arg_value_string(packet, "value");
    if (!name || !value) return ERR_BAD_ARGS;
    if (setenv(name, value, 1) != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

static int exec_sys_exec(OpPacketEx* packet) {
    const char* cmd = arg_value_string(packet, "cmd");
    if (!cmd) return ERR_BAD_ARGS;
    /* DANGEROUS: always requires explicit allow. Validation already checked. */
    int ret = system(cmd);
    if (ret != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

/* --- Build --- */
static int exec_build_compile(OpPacketEx* packet) {
    const char* target = arg_value_string(packet, "target");
    const char* flags = arg_value_string(packet, "flags");
    if (!target) return ERR_BAD_ARGS;
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "make %s", target);
    if (flags) {
        snprintf(cmd, sizeof(cmd), "make %s %s", flags, target);
    }
    int ret = system(cmd);
    if (ret != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

static int exec_build_link(OpPacketEx* packet) {
    const char* inputs = arg_value_string(packet, "inputs");
    const char* output = arg_value_string(packet, "output");
    if (!inputs || !output) return ERR_BAD_ARGS;
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "gcc -o %s %s", output, inputs);
    int ret = system(cmd);
    if (ret != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

static int exec_build_test(OpPacketEx* packet) {
    const char* filter = arg_value_string(packet, "filter");
    int64_t timeout_ms = arg_value_int(packet, "timeout_ms", 30000);
    const char* dir = arg_value_string(packet, "dir");
    (void)timeout_ms;
    char cmd[512];
    if (dir) {
        if (filter) {
            snprintf(cmd, sizeof(cmd), "cd %s && go test -run %s ./...", dir, filter);
        } else {
            snprintf(cmd, sizeof(cmd), "cd %s && go test ./...", dir);
        }
    } else {
        if (filter) {
            snprintf(cmd, sizeof(cmd), "go test -run %s ./...", filter);
        } else {
            snprintf(cmd, sizeof(cmd), "go test ./...");
        }
    }
    int ret = system(cmd);
    if (ret != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

static int exec_build_deploy(OpPacketEx* packet) {
    const char* target = arg_value_string(packet, "target");
    const char* version = arg_value_string(packet, "version");
    if (!target) return ERR_BAD_ARGS;
    char cmd[512];
    if (version) {
        snprintf(cmd, sizeof(cmd), "echo deploy %s@%s", target, version);
    } else {
        snprintf(cmd, sizeof(cmd), "echo deploy %s", target);
    }
    int ret = system(cmd);
    (void)ret;
    /* Deploy is DANGEROUS but we validated permission earlier. */
    return ERR_OK;
}

static int exec_build_clean(OpPacketEx* packet) {
    const char* target = arg_value_string(packet, "target");
    char cmd[512];
    if (target) {
        snprintf(cmd, sizeof(cmd), "make clean-%s", target);
    } else {
        snprintf(cmd, sizeof(cmd), "make clean");
    }
    int ret = system(cmd);
    if (ret != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

/* --- Git stubs (simplified) --- */
static int exec_git_status(OpPacketEx* packet) {
    (void)packet;
    return system("git status --short") == 0 ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_diff(OpPacketEx* packet) {
    (void)packet;
    return system("git diff --stat") == 0 ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_add(OpPacketEx* packet) {
    const char* path = arg_value_string(packet, "path");
    char cmd[512];
    if (path) {
        snprintf(cmd, sizeof(cmd), "git add %s", path);
    } else {
        snprintf(cmd, sizeof(cmd), "git add -A");
    }
    return system(cmd) == 0 ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_commit(OpPacketEx* packet) {
    const char* message = arg_value_string(packet, "message");
    char cmd[512];
    if (message) {
        snprintf(cmd, sizeof(cmd), "git commit -m \"%s\"", message);
    } else {
        snprintf(cmd, sizeof(cmd), "git commit");
    }
    return system(cmd) == 0 ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_checkout(OpPacketEx* packet) {
    const char* branch = arg_value_string(packet, "branch");
    if (!branch) return ERR_BAD_ARGS;
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "git checkout %s", branch);
    return system(cmd) == 0 ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_branch(OpPacketEx* packet) {
    const char* name = arg_value_string(packet, "name");
    char cmd[512];
    if (name) {
        snprintf(cmd, sizeof(cmd), "git branch %s", name);
    } else {
        snprintf(cmd, sizeof(cmd), "git branch --list");
    }
    return system(cmd) == 0 ? ERR_OK : ERR_EXEC_FAIL;
}

/* --- Network stubs --- */
static int exec_net_http_get(OpPacketEx* packet) {
    const char* url = arg_value_string(packet, "url");
    if (!url) return ERR_BAD_ARGS;
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "curl -sL %s", url);
    return system(cmd) == 0 ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_net_http_post(OpPacketEx* packet) {
    const char* url = arg_value_string(packet, "url");
    const char* data = arg_value_string(packet, "data");
    if (!url) return ERR_BAD_ARGS;
    char cmd[512];
    if (data) {
        snprintf(cmd, sizeof(cmd), "curl -sL -X POST -d \"%s\" %s", data, url);
    } else {
        snprintf(cmd, sizeof(cmd), "curl -sL -X POST %s", url);
    }
    return system(cmd) == 0 ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_net_tcp_close(OpPacketEx* packet) {
    int fd = packet->fd_in >= 0 ? packet->fd_in : arg_value_int(packet, "fd", -1);
    if (fd < 0) return ERR_BAD_ARGS;
    if (close(fd) != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

/* --- Process stubs --- */
static int exec_proc_spawn(OpPacketEx* packet) {
    const char* cmd = arg_value_string(packet, "cmd");
    if (!cmd) return ERR_BAD_ARGS;
    /* Simplified: system() blocks. Real spawn is fork+exec. */
    return system(cmd) == 0 ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_proc_wait(OpPacketEx* packet) {
    int64_t pid = arg_value_int(packet, "pid", -1);
    if (pid <= 0) return ERR_BAD_ARGS;
    int status;
    pid_t result = waitpid((pid_t)pid, &status, 0);
    if (result < 0) return ERR_EXEC_FAIL;
    int exit_code = WIFEXITED(status) ? WEXITSTATUS(status) : -1;
    snprintf(packet->result, sizeof(packet->result), "pid=%ld, exit=%d", (long)pid, exit_code);
    packet->result_len = strlen(packet->result);
    return ERR_OK;
}

static int exec_proc_kill(OpPacketEx* packet) {
    int64_t pid = arg_value_int(packet, "pid", -1);
    int64_t sig = arg_value_int(packet, "signal", 9);
    if (pid <= 0) return ERR_BAD_ARGS;
    if (kill((pid_t)pid, (int)sig) != 0) return ERR_EXEC_FAIL;
    return ERR_OK;
}

static int exec_proc_signal(OpPacketEx* packet) {
    return exec_proc_kill(packet); /* Same logic with signal arg */
}

/* --- Utility stubs --- */
static int exec_hash_sha256(OpPacketEx* packet) {
    const char* data = arg_value_string(packet, "data");
    if (!data) return ERR_BAD_ARGS;
    size_t len = strlen(data);
    unsigned char hash[SHA256_DIGEST_LENGTH];
    SHA256((const unsigned char*)data, len, hash);
    char* out = packet->result;
    size_t out_cap = sizeof(packet->result);
    for (int i = 0; i < SHA256_DIGEST_LENGTH && (size_t)(out - packet->result) < out_cap - 3; i++) {
        snprintf(out, out_cap - (out - packet->result), "%02x", hash[i]);
        out += 2;
    }
    *out = 0;
    packet->result_len = (size_t)(out - packet->result);
    return ERR_OK;
}

static int exec_hash_md5(OpPacketEx* packet) {
    const char* data = arg_value_string(packet, "data");
    if (!data) return ERR_BAD_ARGS;
    size_t len = strlen(data);
    unsigned char hash[MD5_DIGEST_LENGTH];
    MD5((const unsigned char*)data, len, hash);
    char* out = packet->result;
    size_t out_cap = sizeof(packet->result);
    for (int i = 0; i < MD5_DIGEST_LENGTH && (size_t)(out - packet->result) < out_cap - 3; i++) {
        snprintf(out, out_cap - (out - packet->result), "%02x", hash[i]);
        out += 2;
    }
    *out = 0;
    packet->result_len = (size_t)(out - packet->result);
    return ERR_OK;
}

static int exec_compress_gzip(OpPacketEx* packet) {
    const char* data = arg_value_string(packet, "data");
    if (!data) return ERR_BAD_ARGS;
    return ERR_OK;
}

static int exec_decompress_gzip(OpPacketEx* packet) {
    const char* data = arg_value_string(packet, "data");
    if (!data) return ERR_BAD_ARGS;
    return ERR_OK;
}

static int exec_encrypt_aes(OpPacketEx* packet) {
    const char* data = arg_value_string(packet, "data");
    if (!data) return ERR_BAD_ARGS;
    return ERR_OK;
}

static int exec_decrypt_aes(OpPacketEx* packet) {
    const char* data = arg_value_string(packet, "data");
    if (!data) return ERR_BAD_ARGS;
    return ERR_OK;
}

/* --- Session / Orchestrator stubs --- */
static int exec_sess_budget_check(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_sess_context_append(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_sess_denial_record(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_sess_snapshot(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_sess_compress(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_orch_classify(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_orch_plan(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_orch_validate(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_orch_exec(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_orch_verify(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_orch_respond(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

/* --- Research stubs --- */
static int exec_research_hypothesis_create(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_hypothesis_load(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_hypothesis_inference(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_experiment_run(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_result_store(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_statistical_test(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_literature_fetch(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_literature_parse(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_literature_index(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_citation_link(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_literature_embed(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_progress_store(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_research_context_summarize(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

/* --- Self-management stubs --- */
static int exec_self_checkpoint_create(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_self_checkpoint_restore(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_self_budget_reallocate(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_self_strategy_pivot(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_self_progress_assess(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

static int exec_self_context_summarize(OpPacketEx* packet) {
    (void)packet;
    return ERR_OK;
}

/* --- Unimplemented placeholder --- */
static int exec_git_init(OpPacketEx* packet) {
    const char* path = arg_value_string(packet, "path");
    if (!path) return ERR_BAD_ARGS;
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "git init %s", path);
    int r = system(cmd);
    snprintf(packet->result, sizeof(packet->result), "git init %s: %s", path, (r == 0) ? "ok" : "fail");
    packet->result_len = strlen(packet->result);
    return (r == 0) ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_clone(OpPacketEx* packet) {
    const char* url = arg_value_string(packet, "url");
    const char* path = arg_value_string(packet, "path");
    if (!url) return ERR_BAD_ARGS;
    char cmd[1024];
    if (path && path[0]) {
        snprintf(cmd, sizeof(cmd), "git clone %s %s", url, path);
    } else {
        snprintf(cmd, sizeof(cmd), "git clone %s", url);
    }
    int r = system(cmd);
    snprintf(packet->result, sizeof(packet->result), "git clone %s: %s", url, (r == 0) ? "ok" : "fail");
    packet->result_len = strlen(packet->result);
    return (r == 0) ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_fetch(OpPacketEx* packet) {
    const char* remote = arg_value_string(packet, "remote");
    if (!remote) remote = "origin";
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "git fetch %s", remote);
    int r = system(cmd);
    snprintf(packet->result, sizeof(packet->result), "git fetch %s: %s", remote, (r == 0) ? "ok" : "fail");
    packet->result_len = strlen(packet->result);
    return (r == 0) ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_push(OpPacketEx* packet) {
    const char* remote = arg_value_string(packet, "remote");
    const char* branch = arg_value_string(packet, "branch");
    if (!remote) remote = "origin";
    if (!branch) branch = "HEAD";
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "git push %s %s", remote, branch);
    int r = system(cmd);
    snprintf(packet->result, sizeof(packet->result), "git push %s %s: %s", remote, branch, (r == 0) ? "ok" : "fail");
    packet->result_len = strlen(packet->result);
    return (r == 0) ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_merge(OpPacketEx* packet) {
    const char* branch = arg_value_string(packet, "branch");
    if (!branch) return ERR_BAD_ARGS;
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "git merge %s", branch);
    int r = system(cmd);
    snprintf(packet->result, sizeof(packet->result), "git merge %s: %s", branch, (r == 0) ? "ok" : "fail");
    packet->result_len = strlen(packet->result);
    return (r == 0) ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_rebase(OpPacketEx* packet) {
    const char* branch = arg_value_string(packet, "branch");
    if (!branch) return ERR_BAD_ARGS;
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "git rebase %s", branch);
    int r = system(cmd);
    snprintf(packet->result, sizeof(packet->result), "git rebase %s: %s", branch, (r == 0) ? "ok" : "fail");
    packet->result_len = strlen(packet->result);
    return (r == 0) ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_tag(OpPacketEx* packet) {
    const char* name = arg_value_string(packet, "name");
    const char* message = arg_value_string(packet, "message");
    if (!name) return ERR_BAD_ARGS;
    char cmd[512];
    if (message && message[0]) {
        snprintf(cmd, sizeof(cmd), "git tag -a %s -m '%s'", name, message);
    } else {
        snprintf(cmd, sizeof(cmd), "git tag %s", name);
    }
    int r = system(cmd);
    snprintf(packet->result, sizeof(packet->result), "git tag %s: %s", name, (r == 0) ? "ok" : "fail");
    packet->result_len = strlen(packet->result);
    return (r == 0) ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_git_reset(OpPacketEx* packet) {
    const char* target = arg_value_string(packet, "target");
    if (!target) target = "HEAD";
    bool hard = arg_value_bool(packet, "hard", false);
    char cmd[512];
    snprintf(cmd, sizeof(cmd), "git reset %s %s", hard ? "--hard" : "", target);
    int r = system(cmd);
    snprintf(packet->result, sizeof(packet->result), "git reset %s: %s", target, (r == 0) ? "ok" : "fail");
    packet->result_len = strlen(packet->result);
    return (r == 0) ? ERR_OK : ERR_EXEC_FAIL;
}

static int exec_unimplemented(OpPacketEx* packet) {
    (void)packet;
    return ERR_EXEC_FAIL;
}

/* ============================================================================
 * Registry builder
 * ============================================================================ */
static void register_op(OpCode opcode, const char* name, const char* desc,
                        int (*exec)(OpPacketEx*),
                        OpCode inv_opcode, int (*inv_exec)(OpPacketEx*, OpPacketEx*),
                        uint32_t req_flags, uint32_t forb_flags,
                        float cost_tok, float cost_time, float cost_mem,
                        uint8_t safety, bool reversible, bool atomic) {
    OpCodeDef def;
    memset(&def, 0, sizeof(def));
    def.opcode = opcode;
    strncpy(def.name, name, OP_NAME_LEN - 1);
    strncpy(def.description, desc, OP_DESC_LEN - 1);
    def.execute = exec;
    def.inverse_opcode = inv_opcode;
    def.inverse_execute = inv_exec;
    def.required_flags = req_flags;
    def.forbidden_flags = forb_flags;
    def.cost_tokens = cost_tok;
    def.cost_time_us = cost_time;
    def.cost_memory_bytes = cost_mem;
    def.safety_level = safety;
    def.is_reversible = reversible;
    def.is_atomic = atomic;
    ops_register(&def);
}

void ops_register_builtins(void) {
    /* NOP */
    register_op(OP_NOP, "NOP", "No operation",
                exec_nop, OP_NOP, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                0.0f, 0.01f, 0.0f,
                3, true, true);

    /* Memory */
    register_op(OP_MMAP_ALLOC, "MMAP_ALLOC", "Allocate anonymous mmap",
                exec_mmap_alloc, OP_MMAP_FREE, NULL,
                OP_FLAG_MEMORY, 0,
                2.0f, 5.0f, 4096.0f,
                2, false, true);
    register_op(OP_MMAP_FREE, "MMAP_FREE", "Free anonymous mmap",
                exec_mmap_free, OP_MMAP_ALLOC, NULL,
                OP_FLAG_MEMORY, 0,
                1.0f, 2.0f, 0.0f,
                2, false, true);
    register_op(OP_MMAP_READ, "MMAP_READ", "Read from mmap",
                exec_mmap_read, OP_MMAP_READ, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 1.0f, 0.0f,
                3, true, true);
    register_op(OP_MMAP_WRITE, "MMAP_WRITE", "Write to mmap",
                exec_mmap_write, OP_MMAP_WRITE, NULL,
                OP_FLAG_MEMORY | OP_FLAG_DISK, 0,
                2.0f, 2.0f, 0.0f,
                2, false, true);
    register_op(OP_MMAP_SYNC, "MMAP_SYNC", "Sync mmap to disk",
                exec_mmap_sync, OP_MMAP_SYNC, NULL,
                OP_FLAG_DISK, 0,
                1.0f, 10.0f, 0.0f,
                2, false, true);

    /* I/O */
    register_op(OP_IO_READ, "IO_READ", "Read from file descriptor",
                exec_io_read, OP_IO_READ, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 5.0f, 0.0f,
                3, true, true);
    register_op(OP_IO_WRITE, "IO_WRITE", "Write to file descriptor",
                exec_io_write, OP_IO_WRITE, NULL,
                OP_FLAG_DISK, 0,
                2.0f, 10.0f, 0.0f,
                2, false, true);
    register_op(OP_IO_OPEN, "IO_OPEN", "Open file",
                exec_io_open, OP_IO_CLOSE, NULL,
                OP_FLAG_DISK, 0,
                2.0f, 50.0f, 0.0f,
                2, false, true);
    register_op(OP_IO_CLOSE, "IO_CLOSE", "Close file descriptor",
                exec_io_close, OP_IO_OPEN, NULL,
                OP_FLAG_DISK, 0,
                1.0f, 5.0f, 0.0f,
                2, false, true);
    register_op(OP_IO_SEEK, "IO_SEEK", "Seek file descriptor",
                exec_io_seek, OP_IO_SEEK, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 1.0f, 0.0f,
                3, true, true);

    /* System */
    register_op(OP_SYS_FILE_EXISTS, "SYS_FILE_EXISTS", "Check file existence",
                exec_sys_file_exists, OP_SYS_FILE_EXISTS, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 10.0f, 0.0f,
                3, true, true);
    register_op(OP_SYS_DIR_CREATE, "SYS_DIR_CREATE", "Create directory",
                exec_sys_dir_create, OP_SYS_DIR_REMOVE, NULL,
                OP_FLAG_DISK, 0,
                2.0f, 50.0f, 4096.0f,
                2, false, true);
    register_op(OP_SYS_DIR_REMOVE, "SYS_DIR_REMOVE", "Remove directory",
                exec_sys_dir_remove, OP_SYS_DIR_CREATE, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                2.0f, 100.0f, 0.0f,
                1, false, false);
    register_op(OP_SYS_FILE_COPY, "SYS_FILE_COPY", "Copy file",
                exec_sys_file_copy, OP_SYS_FILE_DELETE, NULL,
                OP_FLAG_DISK, 0,
                3.0f, 500.0f, 0.0f,
                2, false, true);
    register_op(OP_SYS_FILE_MOVE, "SYS_FILE_MOVE", "Move file",
                exec_sys_file_move, OP_SYS_FILE_MOVE, NULL,
                OP_FLAG_DISK, 0,
                3.0f, 200.0f, 0.0f,
                2, false, true);
    register_op(OP_SYS_FILE_DELETE, "SYS_FILE_DELETE", "Delete file",
                exec_sys_file_delete, OP_SYS_FILE_DELETE, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                1.0f, 20.0f, 0.0f,
                1, false, false);
    register_op(OP_SYS_CHMOD, "SYS_CHMOD", "Change file mode",
                exec_sys_chmod, OP_SYS_CHMOD, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                1.0f, 10.0f, 0.0f,
                1, false, false);
    register_op(OP_SYS_ENV_GET, "SYS_ENV_GET", "Get environment variable",
                exec_sys_env_get, OP_SYS_ENV_GET, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 1.0f, 0.0f,
                3, true, true);
    register_op(OP_SYS_ENV_SET, "SYS_ENV_SET", "Set environment variable",
                exec_sys_env_set, OP_SYS_ENV_SET, NULL,
                OP_FLAG_DANGEROUS, 0,
                1.0f, 2.0f, 0.0f,
                1, false, false);
    register_op(OP_SYS_EXEC, "SYS_EXEC", "Execute shell command",
                exec_sys_exec, OP_SYS_EXEC, NULL,
                OP_FLAG_DANGEROUS, 0,
                5.0f, 10000.0f, 0.0f,
                0, false, false);

    /* Build */
    register_op(OP_BUILD_COMPILE, "BUILD_COMPILE", "Compile target",
                exec_build_compile, OP_BUILD_COMPILE, NULL,
                OP_FLAG_DISK, 0,
                10.0f, 500000.0f, 0.0f,
                2, false, false);
    register_op(OP_BUILD_LINK, "BUILD_LINK", "Link objects",
                exec_build_link, OP_BUILD_LINK, NULL,
                OP_FLAG_DISK, 0,
                5.0f, 100000.0f, 0.0f,
                2, false, false);
    register_op(OP_BUILD_TEST, "BUILD_TEST", "Run tests",
                exec_build_test, OP_BUILD_TEST, NULL,
                OP_FLAG_SAFE, 0,
                10.0f, 300000.0f, 0.0f,
                3, true, false);
    register_op(OP_BUILD_DEPLOY, "BUILD_DEPLOY", "Deploy artifact",
                exec_build_deploy, OP_BUILD_DEPLOY, NULL,
                OP_FLAG_NETWORK | OP_FLAG_DANGEROUS, 0,
                5.0f, 600000.0f, 0.0f,
                0, false, false);
    register_op(OP_BUILD_CLEAN, "BUILD_CLEAN", "Clean build artifacts",
                exec_build_clean, OP_BUILD_CLEAN, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                3.0f, 50000.0f, 0.0f,
                1, false, false);

    /* Git */
    register_op(OP_GIT_STATUS, "GIT_STATUS", "Git status",
                exec_git_status, OP_GIT_STATUS, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                2.0f, 100.0f, 0.0f,
                3, true, true);
    register_op(OP_GIT_DIFF, "GIT_DIFF", "Git diff",
                exec_git_diff, OP_GIT_DIFF, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                3.0f, 200.0f, 0.0f,
                3, true, true);
    register_op(OP_GIT_ADD, "GIT_ADD", "Git add",
                exec_git_add, OP_GIT_ADD, NULL,
                OP_FLAG_DISK, 0,
                2.0f, 50.0f, 0.0f,
                2, false, true);
    register_op(OP_GIT_COMMIT, "GIT_COMMIT", "Git commit",
                exec_git_commit, OP_GIT_COMMIT, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                3.0f, 100.0f, 0.0f,
                0, false, false);
    register_op(OP_GIT_PUSH, "GIT_PUSH", "Git push",
                exec_git_push, OP_GIT_PUSH, NULL,
                OP_FLAG_NETWORK | OP_FLAG_DANGEROUS, 0,
                5.0f, 5000.0f, 0.0f,
                0, false, false);
    register_op(OP_GIT_CHECKOUT, "GIT_CHECKOUT", "Git checkout",
                exec_git_checkout, OP_GIT_CHECKOUT, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                3.0f, 200.0f, 0.0f,
                1, false, false);
    register_op(OP_GIT_BRANCH, "GIT_BRANCH", "Git branch",
                exec_git_branch, OP_GIT_BRANCH, NULL,
                OP_FLAG_DISK, 0,
                2.0f, 50.0f, 0.0f,
                2, false, false);
    register_op(OP_GIT_MERGE, "GIT_MERGE", "Git merge",
                exec_git_merge, OP_GIT_MERGE, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                3.0f, 500.0f, 0.0f,
                0, false, false);
    register_op(OP_GIT_REBASE, "GIT_REBASE", "Git rebase",
                exec_git_rebase, OP_GIT_REBASE, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                3.0f, 500.0f, 0.0f,
                0, false, false);
    register_op(OP_GIT_TAG, "GIT_TAG", "Git tag",
                exec_git_tag, OP_GIT_TAG, NULL,
                OP_FLAG_DISK, 0,
                2.0f, 50.0f, 0.0f,
                1, false, false);
    register_op(OP_GIT_RESET, "GIT_RESET", "Git reset",
                exec_git_reset, OP_GIT_RESET, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                3.0f, 200.0f, 0.0f,
                0, false, false);
    register_op(OP_GIT_INIT, "GIT_INIT", "Git init",
                exec_git_init, OP_GIT_INIT, NULL,
                OP_FLAG_DISK, 0,
                2.0f, 100.0f, 0.0f,
                2, false, false);
    register_op(OP_GIT_CLONE, "GIT_CLONE", "Git clone",
                exec_git_clone, OP_GIT_CLONE, NULL,
                OP_FLAG_NETWORK | OP_FLAG_DISK, 0,
                10.0f, 30000.0f, 0.0f,
                1, false, false);
    register_op(OP_GIT_FETCH, "GIT_FETCH", "Git fetch",
                exec_git_fetch, OP_GIT_FETCH, NULL,
                OP_FLAG_NETWORK, 0,
                5.0f, 10000.0f, 0.0f,
                2, false, false);

    /* Network */
    register_op(OP_NET_HTTP_GET, "NET_HTTP_GET", "HTTP GET",
                exec_net_http_get, OP_NET_HTTP_GET, NULL,
                OP_FLAG_NETWORK, 0,
                5.0f, 5000.0f, 0.0f,
                2, true, true);
    register_op(OP_NET_HTTP_POST, "NET_HTTP_POST", "HTTP POST",
                exec_net_http_post, OP_NET_HTTP_POST, NULL,
                OP_FLAG_NETWORK, 0,
                5.0f, 5000.0f, 0.0f,
                2, false, true);
    register_op(OP_NET_TCP_CONNECT, "NET_TCP_CONNECT", "TCP connect",
                exec_unimplemented, OP_NET_TCP_CLOSE, NULL,
                OP_FLAG_NETWORK, 0,
                3.0f, 2000.0f, 0.0f,
                2, false, false);
    register_op(OP_NET_TCP_SEND, "NET_TCP_SEND", "TCP send",
                exec_unimplemented, OP_NET_TCP_SEND, NULL,
                OP_FLAG_NETWORK, 0,
                2.0f, 1000.0f, 0.0f,
                2, false, true);
    register_op(OP_NET_TCP_RECV, "NET_TCP_RECV", "TCP receive",
                exec_unimplemented, OP_NET_TCP_RECV, NULL,
                OP_FLAG_NETWORK, 0,
                2.0f, 1000.0f, 0.0f,
                2, true, true);
    register_op(OP_NET_TCP_CLOSE, "NET_TCP_CLOSE", "TCP close",
                exec_net_tcp_close, OP_NET_TCP_CONNECT, NULL,
                OP_FLAG_NETWORK, 0,
                1.0f, 50.0f, 0.0f,
                2, false, true);
    register_op(OP_NET_WEBSOCKET, "NET_WEBSOCKET", "WebSocket",
                exec_unimplemented, OP_NET_WEBSOCKET, NULL,
                OP_FLAG_NETWORK, 0,
                5.0f, 3000.0f, 0.0f,
                1, false, false);

    /* Process */
    register_op(OP_PROC_SPAWN, "PROC_SPAWN", "Spawn process",
                exec_proc_spawn, OP_PROC_KILL, NULL,
                OP_FLAG_DANGEROUS, 0,
                5.0f, 10000.0f, 0.0f,
                0, false, false);
    register_op(OP_PROC_WAIT, "PROC_WAIT", "Wait for process",
                exec_proc_wait, OP_PROC_WAIT, NULL,
                OP_FLAG_SAFE, 0,
                2.0f, 5000.0f, 0.0f,
                3, true, true);
    register_op(OP_PROC_KILL, "PROC_KILL", "Kill process",
                exec_proc_kill, OP_PROC_SPAWN, NULL,
                OP_FLAG_DANGEROUS, 0,
                1.0f, 100.0f, 0.0f,
                0, false, false);
    register_op(OP_PROC_SIGNAL, "PROC_SIGNAL", "Signal process",
                exec_proc_signal, OP_PROC_SIGNAL, NULL,
                OP_FLAG_DANGEROUS, 0,
                1.0f, 50.0f, 0.0f,
                1, false, false);

    /* Utility */
    register_op(OP_HASH_SHA256, "HASH_SHA256", "SHA256 hash",
                exec_hash_sha256, OP_HASH_SHA256, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                2.0f, 100.0f, 0.0f,
                3, true, true);
    register_op(OP_HASH_MD5, "HASH_MD5", "MD5 hash",
                exec_hash_md5, OP_HASH_MD5, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 50.0f, 0.0f,
                3, true, true);
    register_op(OP_COMPRESS_GZIP, "COMPRESS_GZIP", "Gzip compress",
                exec_compress_gzip, OP_COMPRESS_GZIP, NULL,
                OP_FLAG_SAFE, 0,
                3.0f, 500.0f, 0.0f,
                3, true, true);
    register_op(OP_DECOMPRESS_GZIP, "DECOMPRESS_GZIP", "Gzip decompress",
                exec_decompress_gzip, OP_DECOMPRESS_GZIP, NULL,
                OP_FLAG_SAFE, 0,
                3.0f, 500.0f, 0.0f,
                3, true, true);
    register_op(OP_ENCRYPT_AES, "ENCRYPT_AES", "AES encrypt",
                exec_encrypt_aes, OP_DECRYPT_AES, NULL,
                OP_FLAG_SAFE, 0,
                3.0f, 200.0f, 0.0f,
                2, true, true);
    register_op(OP_DECRYPT_AES, "DECRYPT_AES", "AES decrypt",
                exec_decrypt_aes, OP_ENCRYPT_AES, NULL,
                OP_FLAG_SAFE, 0,
                3.0f, 200.0f, 0.0f,
                2, true, true);

    /* Session / Orchestrator */
    register_op(OP_SESS_BUDGET_CHECK, "SESS_BUDGET_CHECK", "Check session budget",
                exec_sess_budget_check, OP_SESS_BUDGET_CHECK, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 1.0f, 0.0f,
                3, true, true);
    register_op(OP_SESS_CONTEXT_APPEND, "SESS_CONTEXT_APPEND", "Append to context",
                exec_sess_context_append, OP_SESS_CONTEXT_APPEND, NULL,
                OP_FLAG_MEMORY, 0,
                2.0f, 5.0f, 0.0f,
                2, false, true);
    register_op(OP_SESS_DENIAL_RECORD, "SESS_DENIAL_RECORD", "Record denial",
                exec_sess_denial_record, OP_SESS_DENIAL_RECORD, NULL,
                OP_FLAG_SAFE, 0,
                1.0f, 2.0f, 0.0f,
                3, false, false);
    register_op(OP_SESS_SNAPSHOT, "SESS_SNAPSHOT", "Snapshot session",
                exec_sess_snapshot, OP_SESS_SNAPSHOT, NULL,
                OP_FLAG_DISK, 0,
                5.0f, 100.0f, 0.0f,
                2, false, true);
    register_op(OP_SESS_COMPRESS, "SESS_COMPRESS", "Compress session",
                exec_sess_compress, OP_SESS_COMPRESS, NULL,
                OP_FLAG_SAFE, 0,
                3.0f, 50.0f, 0.0f,
                3, true, true);
    register_op(OP_ORCH_CLASSIFY, "ORCH_CLASSIFY", "Classify request",
                exec_orch_classify, OP_ORCH_CLASSIFY, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                2.0f, 10.0f, 0.0f,
                3, true, true);
    register_op(OP_ORCH_PLAN, "ORCH_PLAN", "Plan execution",
                exec_orch_plan, OP_ORCH_PLAN, NULL,
                OP_FLAG_SAFE, 0,
                3.0f, 20.0f, 0.0f,
                3, false, false);
    register_op(OP_ORCH_VALIDATE, "ORCH_VALIDATE", "Validate chain",
                exec_orch_validate, OP_ORCH_VALIDATE, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                2.0f, 15.0f, 0.0f,
                3, true, true);
    register_op(OP_ORCH_EXEC, "ORCH_EXEC", "Execute plan",
                exec_orch_exec, OP_ORCH_EXEC, NULL,
                OP_FLAG_DANGEROUS, 0,
                5.0f, 100.0f, 0.0f,
                0, false, false);
    register_op(OP_ORCH_VERIFY, "ORCH_VERIFY", "Verify result",
                exec_orch_verify, OP_ORCH_VERIFY, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                2.0f, 15.0f, 0.0f,
                3, true, true);
    register_op(OP_ORCH_RESPOND, "ORCH_RESPOND", "Respond to client",
                exec_orch_respond, OP_ORCH_RESPOND, NULL,
                OP_FLAG_SAFE, 0,
                2.0f, 10.0f, 0.0f,
                3, false, false);

    /* Research */
    register_op(OP_RESEARCH_HYPOTHESIS_CREATE, "RESEARCH_HYPOTHESIS_CREATE", "Create hypothesis",
                exec_research_hypothesis_create, OP_RESEARCH_HYPOTHESIS_CREATE, NULL,
                OP_FLAG_SAFE, 0,
                2.0f, 10.0f, 0.0f,
                3, false, false);
    register_op(OP_RESEARCH_HYPOTHESIS_LOAD, "RESEARCH_HYPOTHESIS_LOAD", "Load hypothesis",
                exec_research_hypothesis_load, OP_RESEARCH_HYPOTHESIS_LOAD, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 5.0f, 0.0f,
                3, true, true);
    register_op(OP_RESEARCH_HYPOTHESIS_INFERENCE, "RESEARCH_HYPOTHESIS_INFERENCE", "Inference on hypothesis",
                exec_research_hypothesis_inference, OP_RESEARCH_HYPOTHESIS_INFERENCE, NULL,
                OP_FLAG_SAFE, 0,
                3.0f, 50.0f, 0.0f,
                3, false, false);
    register_op(OP_RESEARCH_EXPERIMENT_RUN, "RESEARCH_EXPERIMENT_RUN", "Run experiment",
                exec_research_experiment_run, OP_RESEARCH_EXPERIMENT_RUN, NULL,
                OP_FLAG_DANGEROUS, 0,
                10.0f, 60000.0f, 0.0f,
                1, false, false);
    register_op(OP_RESEARCH_RESULT_STORE, "RESEARCH_RESULT_STORE", "Store result",
                exec_research_result_store, OP_RESEARCH_RESULT_STORE, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                3.0f, 50.0f, 0.0f,
                2, false, true);
    register_op(OP_RESEARCH_STATISTICAL_TEST, "RESEARCH_STATISTICAL_TEST", "Statistical test",
                exec_research_statistical_test, OP_RESEARCH_STATISTICAL_TEST, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                3.0f, 100.0f, 0.0f,
                3, true, true);
    register_op(OP_RESEARCH_LITERATURE_FETCH, "RESEARCH_LITERATURE_FETCH", "Fetch literature",
                exec_research_literature_fetch, OP_RESEARCH_LITERATURE_FETCH, NULL,
                OP_FLAG_NETWORK, 0,
                5.0f, 10000.0f, 0.0f,
                2, true, true);
    register_op(OP_RESEARCH_LITERATURE_PARSE, "RESEARCH_LITERATURE_PARSE", "Parse literature",
                exec_research_literature_parse, OP_RESEARCH_LITERATURE_PARSE, NULL,
                OP_FLAG_SAFE, 0,
                5.0f, 500.0f, 0.0f,
                3, true, true);
    register_op(OP_RESEARCH_LITERATURE_INDEX, "RESEARCH_LITERATURE_INDEX", "Index literature",
                exec_research_literature_index, OP_RESEARCH_LITERATURE_INDEX, NULL,
                OP_FLAG_DISK | OP_FLAG_DANGEROUS, 0,
                5.0f, 200.0f, 0.0f,
                2, false, true);
    register_op(OP_RESEARCH_CITATION_LINK, "RESEARCH_CITATION_LINK", "Link citation",
                exec_research_citation_link, OP_RESEARCH_CITATION_LINK, NULL,
                OP_FLAG_SAFE, 0,
                2.0f, 20.0f, 0.0f,
                3, false, false);
    register_op(OP_RESEARCH_LITERATURE_EMBED, "RESEARCH_LITERATURE_EMBED", "Embed literature",
                exec_research_literature_embed, OP_RESEARCH_LITERATURE_EMBED, NULL,
                OP_FLAG_MEMORY, 0,
                10.0f, 5000.0f, 0.0f,
                2, true, true);
    register_op(OP_RESEARCH_PROGRESS_STORE, "RESEARCH_PROGRESS_STORE", "Store progress",
                exec_research_progress_store, OP_RESEARCH_PROGRESS_STORE, NULL,
                OP_FLAG_DISK, 0,
                2.0f, 30.0f, 0.0f,
                2, false, true);
    register_op(OP_RESEARCH_CONTEXT_SUMMARIZE, "RESEARCH_CONTEXT_SUMMARIZE", "Summarize context",
                exec_research_context_summarize, OP_RESEARCH_CONTEXT_SUMMARIZE, NULL,
                OP_FLAG_SAFE, 0,
                5.0f, 200.0f, 0.0f,
                3, true, true);

    /* Self-management */
    register_op(OP_SELF_CHECKPOINT_CREATE, "SELF_CHECKPOINT_CREATE", "Create checkpoint",
                exec_self_checkpoint_create, OP_SELF_CHECKPOINT_CREATE, NULL,
                OP_FLAG_DISK, 0,
                5.0f, 100.0f, 0.0f,
                2, false, true);
    register_op(OP_SELF_CHECKPOINT_RESTORE, "SELF_CHECKPOINT_RESTORE", "Restore checkpoint",
                exec_self_checkpoint_restore, OP_SELF_CHECKPOINT_RESTORE, NULL,
                OP_FLAG_SAFE, 0,
                3.0f, 50.0f, 0.0f,
                3, false, false);
    register_op(OP_SELF_BUDGET_REALLOCATE, "SELF_BUDGET_REALLOCATE", "Reallocate budget",
                exec_self_budget_reallocate, OP_SELF_BUDGET_REALLOCATE, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 5.0f, 0.0f,
                3, true, true);
    register_op(OP_SELF_STRATEGY_PIVOT, "SELF_STRATEGY_PIVOT", "Pivot strategy",
                exec_self_strategy_pivot, OP_SELF_STRATEGY_PIVOT, NULL,
                OP_FLAG_SAFE, 0,
                2.0f, 10.0f, 0.0f,
                3, false, false);
    register_op(OP_SELF_PROGRESS_ASSESS, "SELF_PROGRESS_ASSESS", "Assess progress",
                exec_self_progress_assess, OP_SELF_PROGRESS_ASSESS, NULL,
                OP_FLAG_SAFE | OP_FLAG_READONLY, 0,
                1.0f, 5.0f, 0.0f,
                3, true, true);
    register_op(OP_SELF_CONTEXT_SUMMARIZE, "SELF_CONTEXT_SUMMARIZE", "Summarize context",
                exec_self_context_summarize, OP_SELF_CONTEXT_SUMMARIZE, NULL,
                OP_FLAG_SAFE, 0,
                5.0f, 200.0f, 0.0f,
                3, true, true);
}

/* ============================================================================
 * Execution
 * ============================================================================ */
int ops_execute(OpPacketEx* packet) {
    if (!packet || packet->opcode >= OP_MAX) {
        return ERR_INVALID_OPCODE;
    }
    const OpCodeDef* def = &g_op_registry[packet->opcode];
    if (!def->execute) {
        return ERR_UNREGISTERED;
    }
    return def->execute(packet);
}

int ops_execute_chain(OpPacketEx* packets, size_t count, ExecContext* ctx) {
    if (!packets || count == 0 || !ctx) {
        return ERR_INVALID_CHAIN;
    }

    /* Validate before execution */
    ValidationResult vr = ops_validate_chain(packets, count, ctx);
    if (!vr.is_valid) {
        return ERR_INVALID_CHAIN;
    }

    /* Take pre-state snapshot */
    ctx->pre_state_hash = ops_compute_state_hash(ctx);

    /* Execute */
    for (size_t i = 0; i < count; i++) {
        int result = ops_execute(&packets[i]);
        if (result != ERR_OK) {
            /* Rollback if any preceding op was atomic */
            bool need_rollback = false;
            for (size_t j = 0; j <= i; j++) {
                if (g_op_registry[packets[j].opcode].is_atomic) {
                    need_rollback = true;
                    break;
                }
            }
            if (need_rollback) {
                int rb = ops_rollback_chain(packets, (uint32_t)i, ctx);
                if (rb != ERR_OK) {
                    return ERR_ROLLBACK_FAIL;
                }
            }
            return result;
        }
    }

    return ERR_OK;
}
