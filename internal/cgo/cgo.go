package cgo

/*
#cgo CFLAGS: -I../../core
#cgo LDFLAGS: -L../../core -lcore -lm -lcrypto

#include <stdlib.h>
#include "ops.h"

// Helper functions defined in cgo_helpers.c
extern void cgo_set_arg_string(OpPacketEx* p, size_t idx, const char* key, const char* value);
extern void cgo_set_arg_int(OpPacketEx* p, size_t idx, const char* key, int64_t value);
extern void cgo_set_arg_bool(OpPacketEx* p, size_t idx, const char* key, int value);
extern void cgo_finalize_packet(OpPacketEx* p);
extern OpPacketEx* cgo_alloc_packets(size_t count);
extern void cgo_free_packets(OpPacketEx* arr);
extern const char* cgo_get_error_msg(ValidationResult* vr);
extern ValidationResult cgo_validate_chain(OpPacketEx* packets, size_t count, ExecContext* ctx);
extern int cgo_execute_chain(OpPacketEx* packets, size_t count, ExecContext* ctx);
extern const char* cgo_get_packet_result(OpPacketEx* packets, size_t idx);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Init initializes the c-core. Must be called once before any operation.
func Init() error {
	r := C.ops_init()
	if r != C.ERR_OK {
		return fmt.Errorf("ops_init failed: %d", int(r))
	}
	C.ops_register_builtins()
	return nil
}

// Shutdown cleans up the c-core.
func Shutdown() {
	C.ops_shutdown()
}

// Packet represents a single operation in a chain.
type Packet struct {
	Opcode string
	Args   map[string]interface{}
}

// ChainResult holds the result of chain execution or validation.
type ChainResult struct {
	Valid        bool
	ErrorCode    int
	ErrorMessage string
	EnergyUsed   float32
	LatencyUS    float32
	Result       string // Output from the last packet's execution
}

// EncodeChain converts Go packets to C packets for validation/execution.
func EncodeChain(packets []Packet) unsafe.Pointer {
	count := C.size_t(len(packets))
	cPackets := C.cgo_alloc_packets(count)
	if cPackets == nil {
		return nil
	}

	for i, pkt := range packets {
		opcode := C.ops_string_to_opcode(C.CString(pkt.Opcode))
		cp := (*C.OpPacketEx)(unsafe.Pointer(uintptr(unsafe.Pointer(cPackets)) + uintptr(i)*unsafe.Sizeof(C.OpPacketEx{})))
		C.ops_packet_init(cp, opcode)

		idx := C.size_t(0)
		for k, v := range pkt.Args {
			if idx >= C.MAX_ARGS {
				break
			}
			ck := C.CString(k)
			defer C.free(unsafe.Pointer(ck))
			switch val := v.(type) {
			case string:
				cv := C.CString(val)
				C.cgo_set_arg_string(cp, idx, ck, cv)
				C.free(unsafe.Pointer(cv))
case int, int64, float64:
    var intVal int64
    switch v := val.(type) {
    case int:
        intVal = int64(v)
    case int64:
        intVal = v
    case float64:
        intVal = int64(v)
    }
    C.cgo_set_arg_int(cp, idx, ck, C.int64_t(intVal))
			case bool:
				vint := 0
				if val {
					vint = 1
				}
				C.cgo_set_arg_bool(cp, idx, ck, C.int(vint))
			}
			idx++
		}
		C.cgo_finalize_packet(cp)
	}

	return unsafe.Pointer(cPackets)
}

// FreeChain releases memory allocated by EncodeChain.
func FreeChain(ptr unsafe.Pointer) {
	C.cgo_free_packets((*C.OpPacketEx)(ptr))
}

// ValidateChain checks a chain without executing it.
func ValidateChain(packets []Packet, budgetTokens, budgetTimeMS float32) ChainResult {
	ptr := EncodeChain(packets)
	if ptr == nil {
		return ChainResult{Valid: false, ErrorCode: int(C.ERR_INVALID_CHAIN), ErrorMessage: "failed to allocate packets"}
	}
	defer FreeChain(ptr)

	ctx := &C.ExecContext{}
	ctx.session_budget_tokens = C.float(budgetTokens)
	ctx.session_budget_time_ms = C.float(budgetTimeMS)

	vr := C.cgo_validate_chain((*C.OpPacketEx)(ptr), C.size_t(len(packets)), ctx)

	msg := C.GoString(C.cgo_get_error_msg(&vr))

	return ChainResult{
		Valid:        bool(vr.is_valid),
		ErrorCode:    int(vr.error_code),
		ErrorMessage: msg,
		EnergyUsed:   float32(vr.total_energy),
		LatencyUS:    float32(vr.estimated_latency_us),
	}
}

// ExecuteChain validates and executes a chain of operations.
func ExecuteChain(packets []Packet, budgetTokens, budgetTimeMS float32) (ChainResult, error) {
	ptr := EncodeChain(packets)
	if ptr == nil {
		return ChainResult{Valid: false, ErrorCode: int(C.ERR_INVALID_CHAIN), ErrorMessage: "failed to allocate packets"},
			fmt.Errorf("failed to allocate packets")
	}
	defer FreeChain(ptr)

	ctx := &C.ExecContext{}
	ctx.session_budget_tokens = C.float(budgetTokens)
	ctx.session_budget_time_ms = C.float(budgetTimeMS)

	result := C.cgo_execute_chain((*C.OpPacketEx)(ptr), C.size_t(len(packets)), ctx)

	// Get validation info for energy/latency
	vr := C.cgo_validate_chain((*C.OpPacketEx)(ptr), C.size_t(len(packets)), ctx)

	msg := C.GoString(C.cgo_get_error_msg(&vr))

	cr := ChainResult{
		Valid:        result == C.ERR_OK,
		ErrorCode:    int(result),
		ErrorMessage: msg,
		EnergyUsed:   float32(vr.total_energy),
		LatencyUS:    float32(vr.estimated_latency_us),
	}

	// Capture result from last packet
	if len(packets) > 0 {
		lastIdx := C.size_t(len(packets) - 1)
		cResult := C.cgo_get_packet_result((*C.OpPacketEx)(ptr), lastIdx)
		cr.Result = C.GoString(cResult)
	}

	if result != C.ERR_OK {
		return cr, fmt.Errorf("execution failed: code=%d", int(result))
	}
	return cr, nil
}

// GetAvailableTools returns a list of registered opcodes as tool names.
func GetAvailableTools() []string {
  return []string{
    "SYS_FILE_EXISTS", "SYS_FILE_READ", "SYS_DIR_CREATE", "SYS_DIR_REMOVE",
    "SYS_FILE_COPY", "SYS_FILE_MOVE", "SYS_FILE_DELETE",
    "SYS_CHMOD", "SYS_ENV_GET", "SYS_ENV_SET", "SYS_EXEC",
    "IO_OPEN", "IO_CLOSE", "IO_READ", "IO_WRITE", "IO_SEEK",
    "BUILD_COMPILE", "BUILD_LINK", "BUILD_TEST", "BUILD_DEPLOY", "BUILD_CLEAN",
    "GIT_STATUS", "GIT_DIFF", "GIT_ADD", "GIT_COMMIT", "GIT_CHECKOUT", "GIT_BRANCH",
    "NET_HTTP_GET", "NET_HTTP_POST", "NET_TCP_CLOSE",
    "PROC_SPAWN", "PROC_WAIT", "PROC_KILL", "PROC_SIGNAL",
    "HASH_SHA256", "HASH_MD5",
    "FILE_EDIT", "FILE_INSERT", "FILE_DELETE_RANGE",
    "SESS_BUDGET_CHECK", "ORCH_VALIDATE",
  }
}

// PacketFromToolCall converts an MCP tool call name and arguments into a Packet.
func PacketFromToolCall(name string, args map[string]interface{}) (Packet, error) {
	return Packet{Opcode: name, Args: args}, nil
}
