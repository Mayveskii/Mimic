#include "ops.h"
#include <stdlib.h>
#include <string.h>
#include <stddef.h>

/* ============================================================================
 * CGO Helper Functions
 * ============================================================================ */

void cgo_set_arg_string(OpPacketEx* p, size_t idx, const char* key, const char* value) {
    if (!p || idx >= MAX_ARGS) return;
    strncpy(p->args[idx].key, key, sizeof(p->args[idx].key) - 1);
    p->args[idx].key[sizeof(p->args[idx].key) - 1] = '\0';
    p->args[idx].type = ARG_TYPE_STRING;
    strncpy(p->args[idx].value.s, value, sizeof(p->args[idx].value.s) - 1);
    p->args[idx].value.s[sizeof(p->args[idx].value.s) - 1] = '\0';
}

void cgo_set_arg_int(OpPacketEx* p, size_t idx, const char* key, int64_t value) {
    if (!p || idx >= MAX_ARGS) return;
    strncpy(p->args[idx].key, key, sizeof(p->args[idx].key) - 1);
    p->args[idx].key[sizeof(p->args[idx].key) - 1] = '\0';
    p->args[idx].type = ARG_TYPE_INT;
    p->args[idx].value.i = value;
}

void cgo_set_arg_bool(OpPacketEx* p, size_t idx, const char* key, int value) {
    if (!p || idx >= MAX_ARGS) return;
    strncpy(p->args[idx].key, key, sizeof(p->args[idx].key) - 1);
    p->args[idx].key[sizeof(p->args[idx].key) - 1] = '\0';
    p->args[idx].type = ARG_TYPE_BOOL;
    p->args[idx].value.b = (value != 0);
}

void cgo_finalize_packet(OpPacketEx* p) {
    if (!p) return;
    uint32_t count = 0;
    for (size_t i = 0; i < MAX_ARGS; i++) {
        if (p->args[i].key[0] != '\0') count++;
    }
    p->arg_count = count;
}

OpPacketEx* cgo_alloc_packets(size_t count) {
    OpPacketEx* arr = (OpPacketEx*)calloc(count, sizeof(OpPacketEx));
    return arr;
}

void cgo_free_packets(OpPacketEx* arr) {
    free(arr);
}

const char* cgo_get_error_msg(ValidationResult* vr) {
    if (!vr) return "";
    return vr->error_msg;
}

ValidationResult cgo_validate_chain(OpPacketEx* packets, size_t count, ExecContext* ctx) {
    return ops_validate_chain(packets, count, ctx);
}

int cgo_execute_chain(OpPacketEx* packets, size_t count, ExecContext* ctx) {
    return ops_execute_chain(packets, count, ctx);
}

const char* cgo_get_packet_result(OpPacketEx* packets, size_t idx) {
    if (!packets) return "";
    OpPacketEx* p = &packets[idx];
    return p->result;
}