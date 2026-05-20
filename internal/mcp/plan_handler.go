package mcp

import (
	"fmt"
	"os"

	"github.com/Mayveskii/Mimic/internal/orchestrator"
)

// PlanHandler handles PLAN_GENERATE and PLAN_EXECUTE MCP tools.
type PlanHandler struct {
	budget orchestrator.CostEstimate
}

func NewPlanHandler() *PlanHandler {
	return &PlanHandler{
		budget: orchestrator.CostEstimate{
			Tokens:  100000,
			TimeUs:  10000000,
			MemoryB: 1 << 30,
		},
	}
}

func (h *PlanHandler) HandleGeneratePlan(args map[string]interface{}) map[string]interface{} {
	goal, _ := args["goal"].(string)
	if goal == "" {
		return map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": "Error: 'goal' is required"},
			},
			"isError": true,
		}
	}

	// Build plan (placeholder — real impl would call LLM)
	plan := &orchestrator.Plan{
		ID:     fmt.Sprintf("plan-%d", os.Getpid()),
		Goal:   goal,
		Status: "pending",
	}

	if err := orchestrator.ValidatePlan(plan, h.budget); err != nil {
		return map[string]interface{}{
			"content": []map[string]string{
				{"type": "text", "text": fmt.Sprintf("Plan validation failed: %v", err)},
			},
			"isError": true,
		}
	}

	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": fmt.Sprintf("Plan validated. ID=%s Steps=%d Status=%s", plan.ID, len(plan.Steps), plan.Status)},
		},
	}
}
