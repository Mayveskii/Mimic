/* Auto-generated from decision-patterns.yaml by apply_decisions_to_matrix.py */
#include "decision_matrix.h"

static DecisionConflictRule g_decision_conflicts[] = {
    { OP_BUILD_COMPILE, OP_BUILD_COMPILE, 1, "gonka-ai/vllm", "parallel compiles with torch.compile may contend on Triton compilation cache" },
};

static const size_t g_decision_conflict_count = 1;

static DecisionEnergyCost g_decision_energy[] = {
    { OP_BUILD_COMPILE, 10.0f, 5000000.0f, true, "PR#36: 5ms per kernel fuse" },
};

static const size_t g_decision_energy_count = 1;

void ops_load_decision_conflicts(void) {
    for (size_t i = 0; i < g_decision_conflict_count; i++) {
        DecisionConflictRule* r = &g_decision_conflicts[i];
        g_conflict_matrix[r->op1][r->op2] = r->level;
        g_conflict_matrix[r->op2][r->op1] = r->level;
    }
}

void ops_load_decision_energy(void) {
    for (size_t i = 0; i < g_decision_energy_count; i++) {
        DecisionEnergyCost* c = &g_decision_energy[i];
        g_energy_costs[c->opcode][0] = c->cost_tokens;
        g_energy_costs[c->opcode][1] = c->cost_time_us;
    }
}
