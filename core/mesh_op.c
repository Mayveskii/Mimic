#include "ops.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

/* ============================================================================
 * Mesh Lookup (OP_MESH_QUERY = 0xC0)
 *
 * Bridges to Go mesh lookup via callback:
 * - Go registers callback ops_set_mesh_lookup_fn()
 * - C executor calls it when OP_MESH_QUERY received
 * - Result written to packet->result as JSON
 * ============================================================================ */

/* Mesh lookup function signature (Go callback) */
typedef struct {
    char  query[1024];       /* Natural language query */
    char  domain[64];        /* Optional domain filter */
    int   topK;              /* Number of results to return */
} MeshQueryArgs;

typedef struct {
    char  slot_id[65];       /* SHA256 hex */
    char  domain[64];
    char  invariant[2048];   /* Primary search text */
    char  source_repo[128];
    char  commit[41];
    float survival_index;
    float z_density;
    float similarity;
    char  polarity[16];
} MeshSlotResult;

/* Callback pointer set by Go layer */
static int (*g_mesh_lookup_fn)(const MeshQueryArgs* query, 
                                MeshSlotResult* out_results,
                                int max_results,
                                int* out_count) = NULL;

void ops_set_mesh_lookup_fn(int (*fn)(const MeshQueryArgs*, MeshSlotResult*, int, int*)) {
    g_mesh_lookup_fn = fn;
}

static const char* mesh_get_arg_str(OpPacketEx* packet, const char* key) {
    for (size_t i = 0; i < packet->arg_count; i++) {
        if (strcmp(packet->args[i].key, key) == 0 && packet->args[i].type == ARG_TYPE_STRING) {
            return packet->args[i].value.s;
        }
    }
    return NULL;
}

/* Execute OP_MESH_QUERY */
int exec_mesh_query(OpPacketEx* packet) {
    const char* query_text = mesh_get_arg_str(packet, "query");
    const char* domain = mesh_get_arg_str(packet, "domain");
    const char* topK_str = mesh_get_arg_str(packet, "topK");
    int topK = 5;
    
    if (!query_text || strlen(query_text) == 0) {
        snprintf(packet->result, sizeof(packet->result), 
                 "{\"error\":\"mesh_query requires 'query' argument\"}");
        return ERR_BAD_ARGS;
    }
    
    if (!g_mesh_lookup_fn) {
        snprintf(packet->result, sizeof(packet->result),
                 "{\"error\":\"mesh lookup not initialized\"}");
        return ERR_NOT_INIT;
    }
    
    if (topK_str) {
        topK = atoi(topK_str);
    }
    if (topK <= 0) topK = 5;
    if (topK > 20) topK = 20;
    
    MeshQueryArgs args;
    memset(&args, 0, sizeof(args));
    strncpy(args.query, query_text, sizeof(args.query) - 1);
    if (domain) {
        strncpy(args.domain, domain, sizeof(args.domain) - 1);
    }
    args.topK = topK;
    
    MeshSlotResult results[20];
    memset(results, 0, sizeof(results));
    int result_count = 0;
    
    int rc = g_mesh_lookup_fn(&args, results, topK, &result_count);
    if (rc != 0 || result_count == 0) {
        snprintf(packet->result, sizeof(packet->result),
                 "{\"status\":\"no_results\",\"query\":\"%s\",\"reason\":\"no matching patterns found\"}",
                 args.query);
        return ERR_OK;
    }
    
    char* p = packet->result;
    size_t remain = sizeof(packet->result);
    int n = snprintf(p, remain, "{\"status\":\"success\",\"query\":\"%s\",\"results\":[", args.query);
    p += n; remain -= n;
    
    for (int i = 0; i < result_count && remain > 256; i++) {
        n = snprintf(p, remain,
            "%s{\"slot_id\":\"%s\",\"domain\":\"%s\",\"similarity\":%.3f,"
            "\"survival_index\":%.2f,\"z_density\":%.2f,\"source_repo\":\"%s\","
            "\"invariant\":\"%.128s\"}",
            (i > 0) ? "," : "",
            results[i].slot_id,
            results[i].domain,
            results[i].similarity,
            results[i].survival_index,
            results[i].z_density,
            results[i].source_repo,
            results[i].invariant);
        p += n; remain -= n;
    }
    
    if (remain > 10) {
        snprintf(p, remain, "]}");
    }
    
    packet->result_len = strlen(packet->result);
    return ERR_OK;
}
