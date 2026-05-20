#include "ops.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

/* ============================================================================
 * Pattern Execution (OP_EXECUTE_PATTERN = 0xD0)
 *
 * Executes ActionBytes from a mesh slot.
 * ActionBytes format (heuristic):
 *   - If starts with '[' or '{': JSON array of command objects
 *   - Otherwise: shell command string executed via popen
 *
 * Future: full binary OpPacket chain deserialization from embryo format.
 * ============================================================================ */

static const char* pattern_get_arg_str(OpPacketEx* packet, const char* key) {
    for (size_t i = 0; i < packet->arg_count; i++) {
        if (strcmp(packet->args[i].key, key) == 0 && packet->args[i].type == ARG_TYPE_STRING) {
            return packet->args[i].value.s;
        }
    }
    return NULL;
}

/* Simple JSON array parser for command strings: ["cmd1", "cmd2", ...] */
static int parse_json_string_array(const char* json, char** out_cmds, int max_cmds) {
    int count = 0;
    const char* p = json;
    while (*p && count < max_cmds) {
        /* Find quoted string */
        while (*p && *p != '"') p++;
        if (!*p) break;
        p++; /* skip opening quote */
        const char* start = p;
        while (*p && *p != '"') p++;
        if (!*p) break;
        int len = p - start;
        if (len > 0 && len < 1024) {
            out_cmds[count] = malloc(len + 1);
            if (out_cmds[count]) {
                strncpy(out_cmds[count], start, len);
                out_cmds[count][len] = '\0';
                count++;
            }
        }
        p++; /* skip closing quote */
        while (*p && *p != ',' && *p != ']') p++;
    }
    return count;
}

/* Execute a single shell command and capture output */
static int exec_shell_capture(const char* cmd, char* out_buf, size_t out_size) {
    FILE* fp = popen(cmd, "r");
    if (!fp) {
        snprintf(out_buf, out_size, "{\"error\":\"popen failed for: %.64s\"}", cmd);
        return ERR_EXEC_FAIL;
    }
    size_t n = fread(out_buf, 1, out_size - 1, fp);
    out_buf[n] = '\0';
    int status = pclose(fp);
    if (status != 0) {
        char tmp[ERR_MSG_LEN];
        snprintf(tmp, sizeof(tmp), "{\"status\":\"nonzero_exit\",\"exit_code\":%d,\"partial_output\":\"%.128s\"}",
                 WEXITSTATUS(status), out_buf);
        strncpy(out_buf, tmp, out_size - 1);
        out_buf[out_size - 1] = '\0';
        return ERR_EXEC_FAIL;
    }
    return ERR_OK;
}

/* Execute OP_EXECUTE_PATTERN */
int exec_execute_pattern(OpPacketEx* packet) {
    const char* action_bytes = pattern_get_arg_str(packet, "action_bytes");
    const char* slot_id = pattern_get_arg_str(packet, "slot_id");
    
    if (!action_bytes || strlen(action_bytes) == 0) {
        snprintf(packet->result, sizeof(packet->result),
                 "{\"error\":\"EXECUTE_PATTERN requires 'action_bytes'\"}");
        return ERR_BAD_ARGS;
    }
    
    char* p = packet->result;
    size_t remain = sizeof(packet->result);
    int n = snprintf(p, remain, "{\"slot_id\":\"%s\",\"executed\":true,\"output\":\"", slot_id ? slot_id : "unknown");
    p += n; remain -= n;
    
    /* Trim leading whitespace */
    while (*action_bytes && isspace((unsigned char)*action_bytes)) action_bytes++;
    
    int rc = ERR_OK;
    if (*action_bytes == '[' || *action_bytes == '{') {
        /* Try JSON array of commands */
        char* cmds[16];
        int cmd_count = parse_json_string_array(action_bytes, cmds, 16);
        if (cmd_count > 0) {
            for (int i = 0; i < cmd_count && remain > 256; i++) {
                char buf[1024];
                int sub_rc = exec_shell_capture(cmds[i], buf, sizeof(buf));
                int out_len = snprintf(p, remain, "[cmd %d: %s] %.256s\\n", i+1, 
                                       sub_rc == 0 ? "OK" : "FAIL", buf);
                p += out_len; remain -= out_len;
                if (sub_rc != 0) rc = sub_rc;
                free(cmds[i]);
            }
        } else {
            n = snprintf(p, remain, "{\"error\":\"failed to parse JSON array\"}");
            p += n; remain -= n;
            rc = ERR_BAD_ARGS;
        }
    } else if (action_bytes[0] == '!' || (unsigned char)action_bytes[0] < 32) {
        /* Binary patch format (embryo ActionBytes) — return as base64 for agent analysis */
        n = snprintf(p, remain, "[BINARY_PATCH_FORMAT] ActionBytes is an embryo binary patch (not executable as shell). Length=%zu bytes. First 32 bytes hex: ", strlen(action_bytes));
        p += n; remain -= n;
        for (int i = 0; i < 32 && i < (int)strlen(action_bytes) && remain > 4; i++) {
            n = snprintf(p, remain, "%02x", (unsigned char)action_bytes[i]);
            p += n; remain -= n;
        }
        n = snprintf(p, remain, "...");
        p += n; remain -= n;
        rc = ERR_OK;
    } else {
        /* Treat as shell command */
        char buf[4096];
        rc = exec_shell_capture(action_bytes, buf, sizeof(buf));
        int out_len = snprintf(p, remain, "%.1024s", buf);
        p += out_len; remain -= out_len;
    }
    
    if (remain > 4) {
        snprintf(p, remain, "\"}");
    }
    
    packet->result_len = strlen(packet->result);
    return rc;
}
