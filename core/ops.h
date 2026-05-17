#ifndef MIMIC_OPS_H
#define MIMIC_OPS_H

#include <stdint.h>
#include <stddef.h>

typedef struct {
    uint8_t opcode;
    uint8_t flags;
    uint16_t slot;
    uint32_t size;
    const void *data;
} OpPacket;

typedef struct {
    uint8_t opcode;
    const char *name;
    uint8_t min_args;
    uint8_t max_args;
    uint16_t flags;
} OpCodeDef;

int ops_execute_chain(const OpPacket *packets, size_t count);

#endif
