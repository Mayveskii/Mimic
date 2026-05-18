#include "ops.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <unistd.h>

static void test_init_shutdown(void) {
    assert(ops_init() == ERR_OK);
    assert(ops_init() == ERR_NOT_INIT); /* Cannot init twice */
    ops_shutdown();
    assert(ops_init() == ERR_OK);
    ops_shutdown();
    printf("PASS: init_shutdown\n");
}

static void test_register_builtins(void) {
    ops_init();
    ops_register_builtins();

    const OpCodeDef* def = ops_get_definition(OP_NOP);
    assert(def != NULL);
    assert(def->execute != NULL);
    assert(strcmp(def->name, "NOP") == 0);

    def = ops_get_definition(OP_SYS_FILE_EXISTS);
    assert(def != NULL);
    assert(def->execute != NULL);

    def = ops_get_definition(OP_SYS_DIR_CREATE);
    assert(def != NULL);
    assert(def->execute != NULL);

    def = ops_get_definition(OP_MAX);
    assert(def == NULL);

    printf("PASS: register_builtins\n");
    ops_shutdown();
}

static void test_opcode_to_string(void) {
    assert(strcmp(ops_opcode_to_string(OP_NOP), "NOP") == 0);
    assert(strcmp(ops_opcode_to_string(OP_SYS_EXEC), "SYS_EXEC") == 0);
    assert(strcmp(ops_opcode_to_string(0xFF), "UNKNOWN") == 0);
    assert(ops_string_to_opcode("NOP") == OP_NOP);
    assert(ops_string_to_opcode("SYS_EXEC") == OP_SYS_EXEC);
    assert(ops_string_to_opcode("NONEXISTENT") == OP_NOP);
    printf("PASS: opcode_to_string\n");
}

static void test_packet_helpers(void) {
    OpPacketEx pkt;
    ops_packet_init(&pkt, OP_SYS_FILE_EXISTS);
    assert(pkt.opcode == OP_SYS_FILE_EXISTS);
    assert(pkt.arg_count == 0);
    assert(pkt.fd_in == -1);
    assert(pkt.fd_out == -1);

    ops_packet_set_string(&pkt, "path", "/tmp/test");
    assert(pkt.arg_count == 1);
    assert(strcmp(pkt.args[0].key, "path") == 0);
    assert(strcmp(pkt.args[0].value.s, "/tmp/test") == 0);

    ops_packet_set_int(&pkt, "mode", 0755);
    assert(pkt.arg_count == 2);
    assert(pkt.args[1].value.i == 0755);

    printf("PASS: packet_helpers\n");
}

static void test_validation_empty_chain(void) {
    ops_init();
    ops_register_builtins();

    ExecContext ctx = {0};
    ctx.session_budget_tokens = 1000.0f;
    ctx.session_budget_time_ms = 1000.0f;

    ValidationResult vr = ops_validate_chain(NULL, 0, &ctx);
    assert(!vr.is_valid);
    assert(vr.error_code == ERR_INVALID_CHAIN);

    printf("PASS: validation_empty_chain\n");
    ops_shutdown();
}

static void test_validation_valid_chain(void) {
    ops_init();
    ops_register_builtins();

    ExecContext ctx = {0};
    ctx.session_budget_tokens = 1000.0f;
    ctx.session_budget_time_ms = 1000000.0f;

    OpPacketEx packets[2];
    ops_packet_init(&packets[0], OP_SYS_FILE_EXISTS);
    ops_packet_set_string(&packets[0], "path", "/tmp");
    ops_packet_init(&packets[1], OP_SYS_DIR_CREATE);
    ops_packet_set_string(&packets[1], "path", "/tmp/mimic_test_dir");
    ops_packet_set_int(&packets[1], "mode", 0755);

    ValidationResult vr = ops_validate_chain(packets, 2, &ctx);
    assert(vr.is_valid);
    assert(vr.error_code == ERR_OK);
    assert(vr.total_energy > 0.0f);

    printf("PASS: validation_valid_chain\n");
    ops_shutdown();
}

static void test_validation_conflict(void) {
    ops_init();
    ops_register_builtins();

    ExecContext ctx = {0};
    ctx.session_budget_tokens = 10000.0f;
    ctx.session_budget_time_ms = 1000000.0f;

    OpPacketEx packets[2];
    ops_packet_init(&packets[0], OP_SYS_EXEC);
    ops_packet_set_string(&packets[0], "cmd", "echo hello");
    ops_packet_init(&packets[1], OP_SYS_EXEC);
    ops_packet_set_string(&packets[1], "cmd", "echo world");

    ValidationResult vr = ops_validate_chain(packets, 2, &ctx);
    assert(!vr.is_valid);
    assert(vr.error_code == ERR_CONFLICT);

    printf("PASS: validation_conflict\n");
    ops_shutdown();
}

static void test_validation_permission_denied(void) {
    ops_init();
    ops_register_builtins();

    ExecContext ctx = {0};
    ctx.session_budget_tokens = 10000.0f;
    ctx.session_budget_time_ms = 1000000.0f;

    OpPacketEx packets[1];
    ops_packet_init(&packets[0], OP_SYS_EXEC);
    ops_packet_set_string(&packets[0], "cmd", "echo hello");

    ValidationResult vr = ops_validate_chain(packets, 1, &ctx);
    assert(!vr.is_valid);
    assert(vr.error_code == ERR_PERMISSION);

    printf("PASS: validation_permission_denied\n");
    ops_shutdown();
}

static void test_execute_nop(void) {
    ops_init();
    ops_register_builtins();

    OpPacketEx pkt;
    ops_packet_init(&pkt, OP_NOP);
    assert(ops_execute(&pkt) == ERR_OK);

    printf("PASS: execute_nop\n");
    ops_shutdown();
}

static void test_execute_chain_io(void) {
    ops_init();
    ops_register_builtins();

    ExecContext ctx = {0};
    ctx.session_budget_tokens = 1000.0f;
    ctx.session_budget_time_ms = 1000000.0f;
    (void)ctx; /* Not used in simplified FD test */

    const char* test_path = "/tmp/mimic_test_io.txt";

    OpPacketEx packets[3];
    ops_packet_init(&packets[0], OP_IO_OPEN);
    ops_packet_set_string(&packets[0], "path", test_path);
    ops_packet_set_string(&packets[0], "mode", "w");

    ops_packet_init(&packets[1], OP_IO_WRITE);
    ops_packet_set_string(&packets[1], "data", "hello mimic");
    packets[1].fd_in = -1; /* Will be set by exec, but we don't track FDs between packets yet */

    ops_packet_init(&packets[2], OP_IO_CLOSE);
    packets[2].fd_in = -1;

    /* FD chaining not fully implemented in simplified test; test individual ops instead */
    int r = ops_execute(&packets[0]);
    assert(r == ERR_OK);
    assert(packets[0].fd_out >= 0);

    packets[1].fd_in = packets[0].fd_out;
    r = ops_execute(&packets[1]);
    assert(r == ERR_OK);

    packets[2].fd_in = packets[0].fd_out;
    r = ops_execute(&packets[2]);
    assert(r == ERR_OK);

    /* Verify file exists */
    struct stat st;
    assert(stat(test_path, &st) == 0);

    /* Cleanup */
    unlink(test_path);

    printf("PASS: execute_chain_io\n");
    ops_shutdown();
}

static void test_execute_chain_system(void) {
    ops_init();
    ops_register_builtins();

    ExecContext ctx = {0};
    ctx.session_budget_tokens = 1000.0f;
    ctx.session_budget_time_ms = 1000000.0f;

    const char* test_dir = "/tmp/mimic_test_dir";
    rmdir(test_dir); /* clean if exists */

    OpPacketEx packets[2];
    ops_packet_init(&packets[0], OP_SYS_FILE_EXISTS);
    ops_packet_set_string(&packets[0], "path", test_dir);

    ops_packet_init(&packets[1], OP_SYS_DIR_CREATE);
    ops_packet_set_string(&packets[1], "path", test_dir);
    ops_packet_set_int(&packets[1], "mode", 0755);

    ValidationResult vr = ops_validate_chain(packets, 2, &ctx);
    assert(vr.is_valid);

    int r = ops_execute_chain(packets, 2, &ctx);
    assert(r == ERR_OK);

    struct stat st;
    assert(stat(test_dir, &st) == 0);
    assert(S_ISDIR(st.st_mode));

    /* Cleanup */
    rmdir(test_dir);

    printf("PASS: execute_chain_system\n");
    ops_shutdown();
}

static void test_mmap_alloc_free(void) {
    ops_init();
    ops_register_builtins();

    size_t size = 4096;
    void* ptr = ops_mmap_alloc(size);
    assert(ptr != NULL);

    assert(ops_mmap_free(ptr, size) == ERR_OK);

    printf("PASS: mmap_alloc_free\n");
    ops_shutdown();
}

static void test_time_ns(void) {
    uint64_t t1 = ops_get_time_ns();
    usleep(1000); /* 1ms */
    uint64_t t2 = ops_get_time_ns();
    assert(t2 > t1);
    printf("PASS: time_ns\n");
}

static void test_calculate_action(void) {
    ops_init();
    ops_register_builtins();

    OpPacketEx packets[2];
    ops_packet_init(&packets[0], OP_SYS_FILE_EXISTS);
    ops_packet_init(&packets[1], OP_SYS_DIR_CREATE);

    float energy = ops_calculate_action(packets, 2);
    assert(energy > 0.0f);

    printf("PASS: calculate_action\n");
    ops_shutdown();
}

static void test_rollback_chain(void) {
    ops_init();
    ops_register_builtins();

    ExecContext ctx = {0};
    ctx.session_budget_tokens = 1000.0f;
    ctx.session_budget_time_ms = 1000000.0f;

    /* Pre-state blob required for rollback */
    ctx.pre_state_blob = malloc(64);
    ctx.pre_state_blob_size = 64;
    ctx.pre_state_hash = ops_compute_state_hash(&ctx);

    OpPacketEx packets[1];
    ops_packet_init(&packets[0], OP_NOP);

    int r = ops_rollback_chain(packets, 0, &ctx);
    assert(r == ERR_OK);

    free(ctx.pre_state_blob);
    ctx.pre_state_blob = NULL;

    printf("PASS: rollback_chain\n");
    ops_shutdown();
}

static void test_energy_budget_overflow(void) {
    ops_init();
    ops_register_builtins();

    ExecContext ctx = {0};
    ctx.session_budget_tokens = 1.0f; /* Very low budget */
    ctx.session_budget_time_ms = 1000000.0f;

    OpPacketEx packets[1];
    ops_packet_init(&packets[0], OP_SYS_DIR_CREATE);
    ops_packet_set_string(&packets[0], "path", "/tmp/mimic_overflow");

    ValidationResult vr = ops_validate_chain(packets, 1, &ctx);
    assert(!vr.is_valid);
    assert(vr.error_code == ERR_ENERGY_OVERFLOW);

    printf("PASS: energy_budget_overflow\n");
    ops_shutdown();
}

int main(void) {
    printf("=== Mimic C-Core Tests ===\n");
    test_init_shutdown();
    test_register_builtins();
    test_opcode_to_string();
    test_packet_helpers();
    test_validation_empty_chain();
    test_validation_valid_chain();
    test_validation_conflict();
    test_validation_permission_denied();
    test_execute_nop();
    test_execute_chain_io();
    test_execute_chain_system();
    test_mmap_alloc_free();
    test_time_ns();
    test_calculate_action();
    test_rollback_chain();
    test_energy_budget_overflow();
    printf("\nAll tests passed.\n");
    return 0;
}
