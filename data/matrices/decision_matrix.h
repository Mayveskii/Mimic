/* Auto-generated from decision-patterns.yaml by apply_decisions_to_matrix.py */
#ifndef MIMIC_DECISION_MATRIX_H
#define MIMIC_DECISION_MATRIX_H

#include "ops.h"

/* Decision-derived conflict rules */
typedef struct {
    OpCode op1;
    OpCode op2;
    ConflictLevel level;
    const char* source;
    const char* rationale;
} DecisionConflictRule;

/* Decision-derived energy costs */
typedef struct {
    OpCode opcode;
    float cost_tokens;
    float cost_time_us;
    bool measured;
    const char* source;
} DecisionEnergyCost;

void ops_load_decision_conflicts(void);
void ops_load_decision_energy(void);

#endif
