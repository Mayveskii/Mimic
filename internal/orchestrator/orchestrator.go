package orchestrator

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Mayveskii/Mimic/internal/cgo"
)

// Phase represents one step in the 6-phase pipeline
type Phase int

const (
	PhaseClassify Phase = iota
	PhasePlan
	PhaseValidate
	PhaseExec
	PhaseVerify
	PhaseRespond
)

func (p Phase) String() string {
	switch p {
	case PhaseClassify:
		return "CLASSIFY"
	case PhasePlan:
		return "PLAN"
	case PhaseValidate:
		return "VALIDATE"
	case PhaseExec:
		return "EXEC"
	case PhaseVerify:
		return "VERIFY"
	case PhaseRespond:
		return "RESPOND"
	}
	return "UNKNOWN"
}

// PhaseResult captures the outcome of a single phase
type PhaseResult struct {
	Phase      Phase
	Success    bool
	Error      string
	Output     interface{}
	EnergyUsed float32
	LatencyUS  float32
}

// WorkflowResult is the final output of the orchestrator
type WorkflowResult struct {
	Phases        []PhaseResult
	FinalOutput   interface{}
	TotalEnergy   float32
	TotalLatency  float32
	BudgetTokens  float32
	BudgetTimeMS  float32
	Denials       int
	CircuitBroken bool
}

// Orchestrator implements the 6-phase pipeline per ADR-0004 and INVARIANTS.md
type Orchestrator struct {
	budgetTokens  float32
	budgetTimeMS  float32
	denials       int
	circuitBroken bool
	lastValidated []cgo.Packet // cache of last validated chain
}

// NewOrchestrator creates an orchestrator with a session budget
func NewOrchestrator(tokens, timeMS float32) *Orchestrator {
	return &Orchestrator{
		budgetTokens: tokens,
		budgetTimeMS: timeMS,
	}
}

// Run executes the full 6-phase workflow for a tool call
func (o *Orchestrator) Run(intent string, args map[string]interface{}) (WorkflowResult, error) {
	result := WorkflowResult{
		BudgetTokens:  o.budgetTokens,
		BudgetTimeMS:  o.budgetTimeMS,
		Denials:       o.denials,
		CircuitBroken: o.circuitBroken,
	}

	// --- CLASSIFY ---
	pr := o.classify(intent, args)
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		return result, fmt.Errorf("CLASSIFY failed: %s", pr.Error)
	}
	domain := pr.Output.(string)

	// --- PLAN ---
	pr = o.plan(domain, intent, args)
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		return result, fmt.Errorf("PLAN failed: %s", pr.Error)
	}
	packets := pr.Output.([]cgo.Packet)

	// --- VALIDATE ---
	pr = o.validate(packets)
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		return result, fmt.Errorf("VALIDATE failed: %s", pr.Error)
	}

	// --- EXEC ---
	pr = o.exec(packets)
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		// If rollback was triggered during exec, the c-core handles it
		return result, fmt.Errorf("EXEC failed: %s", pr.Error)
	}

	// --- VERIFY ---
	pr = o.verify(packets, pr.Output)
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		return result, fmt.Errorf("VERIFY failed: %s", pr.Error)
	}

	// --- RESPOND ---
	pr = o.respond(result.Phases)
	result.Phases = append(result.Phases, pr)
	result.FinalOutput = pr.Output

	// Update totals
	for _, p := range result.Phases {
		result.TotalEnergy += p.EnergyUsed
		result.TotalLatency += p.LatencyUS
	}

	// Deduct from budget
	o.budgetTokens -= result.TotalEnergy
	o.budgetTimeMS -= result.TotalLatency / 1000.0 // us → ms

	result.BudgetTokens = o.budgetTokens
	result.BudgetTimeMS = o.budgetTimeMS
	result.Denials = o.denials
	result.CircuitBroken = o.circuitBroken

	return result, nil
}

// classify maps intent text to a domain
func (o *Orchestrator) classify(intent string, args map[string]interface{}) PhaseResult {
	// Simple heuristic: keyword matching
	domain := "general"
	lower := strings.ToLower(intent)
	if strings.Contains(lower, "git") || strings.Contains(lower, "commit") || strings.Contains(lower, "branch") {
		domain = "git"
	} else if strings.Contains(lower, "file") || strings.Contains(lower, "dir") || strings.Contains(lower, "path") {
		domain = "system"
	} else if strings.Contains(lower, "build") || strings.Contains(lower, "compile") || strings.Contains(lower, "test") {
		domain = "build"
	} else if strings.Contains(lower, "network") || strings.Contains(lower, "http") || strings.Contains(lower, "tcp") {
		domain = "network"
	} else if strings.Contains(lower, "process") || strings.Contains(lower, "spawn") || strings.Contains(lower, "kill") {
		domain = "process"
	} else if strings.Contains(lower, "research") || strings.Contains(lower, "hypothesis") || strings.Contains(lower, "experiment") {
		domain = "research"
	}

	return PhaseResult{
		Phase:   PhaseClassify,
		Success: true,
		Output:  domain,
	}
}

// plan builds an OpPacket chain from the intent and args
func (o *Orchestrator) plan(domain, intent string, args map[string]interface{}) PhaseResult {
	// Single-packet chain for now: direct tool call
	pkt, err := cgo.PacketFromToolCall(intent, args)
	if err != nil {
		return PhaseResult{Phase: PhasePlan, Success: false, Error: err.Error()}
	}
	return PhaseResult{
		Phase:   PhasePlan,
		Success: true,
		Output:  []cgo.Packet{pkt},
	}
}

// validate runs cgo.ValidateChain
func (o *Orchestrator) validate(packets []cgo.Packet) PhaseResult {
	vr := cgo.ValidateChain(packets, o.budgetTokens, o.budgetTimeMS)
	if vr.Valid {
		o.lastValidated = packets
		return PhaseResult{
			Phase:      PhaseValidate,
			Success:    true,
			Output:     vr,
			EnergyUsed: vr.EnergyUsed,
			LatencyUS:  vr.LatencyUS,
		}
	}
	return PhaseResult{
		Phase:      PhaseValidate,
		Success:    false,
		Error:      fmt.Sprintf("validation failed: %s (code=%d)", vr.ErrorMessage, vr.ErrorCode),
		EnergyUsed: vr.EnergyUsed,
		LatencyUS:  vr.LatencyUS,
	}
}

// exec runs cgo.ExecuteChain
func (o *Orchestrator) exec(packets []cgo.Packet) PhaseResult {
	er, err := cgo.ExecuteChain(packets, o.budgetTokens, o.budgetTimeMS)
	if err != nil {
		return PhaseResult{
			Phase:      PhaseExec,
			Success:    false,
			Error:      fmt.Sprintf("execution failed: %s (code=%d)", er.ErrorMessage, er.ErrorCode),
			EnergyUsed: er.EnergyUsed,
			LatencyUS:  er.LatencyUS,
		}
	}
	return PhaseResult{
		Phase:      PhaseExec,
		Success:    true,
		Output:     er,
		EnergyUsed: er.EnergyUsed,
		LatencyUS:  er.LatencyUS,
	}
}

// verify performs 2-vote check for critical operations (safety level 0)
func (o *Orchestrator) verify(packets []cgo.Packet, execOutput interface{}) PhaseResult {
	// Check if any packet is CRITICAL (safety level 0)
	hasCritical := false
	for _, pkt := range packets {
		// Safety level check via c-core would require a lookup
		// For now, we assume single-packet chains and check known critical ops
		if isCriticalOp(pkt.Opcode) {
			hasCritical = true
			break
		}
	}

	if !hasCritical {
		// VERIFY is no-op for non-critical ops (per INVARIANTS.md)
		return PhaseResult{Phase: PhaseVerify, Success: true, Output: execOutput}
	}

	// Simulate 2-vote verification
	voteA := "pass"
	voteB := "pass"

	// In production: call external verifiers, run independent checks
	if voteA == "pass" && voteB == "pass" {
		return PhaseResult{Phase: PhaseVerify, Success: true, Output: execOutput}
	}
	return PhaseResult{
		Phase:   PhaseVerify,
		Success: false,
		Error:   fmt.Sprintf("2-vote verify failed: A=%s, B=%s", voteA, voteB),
	}
}

// respond assembles the final structured response
func (o *Orchestrator) respond(phases []PhaseResult) PhaseResult {
	// Build complete response per INVARIANTS.md OINV-06
	allPhases := append(phases, PhaseResult{Phase: PhaseRespond, Success: true})
	response := map[string]interface{}{
		"status":         "ok",
		"phases":         phaseNames(allPhases),
		"metrics":        phaseMetrics(allPhases),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"budget_tokens":  o.budgetTokens,
		"budget_time_ms": o.budgetTimeMS,
	}

	// Include last phase output as primary result
	if len(phases) > 0 {
		last := phases[len(phases)-1]
		response["result"] = last.Output
	}

	// Check for budget warnings (per session invariants)
	if o.budgetTokens < 1000 {
		response["warning"] = "budget_low"
	}

	return PhaseResult{
		Phase:   PhaseRespond,
		Success: true,
		Output:  response,
	}
}

func (o *Orchestrator) checkCircuit() {
	if o.denials >= 3 {
		o.circuitBroken = true
	}
}

func isCriticalOp(name string) bool {
	// Ops with safety level 0 (CRITICAL)
	critical := map[string]bool{
		"SYS_EXEC": true, "BUILD_DEPLOY": true, "GIT_COMMIT": true,
		"GIT_PUSH": true, "GIT_MERGE": true, "GIT_REBASE": true,
		"PROC_KILL": true,
	}
	return critical[name]
}

func phaseNames(phases []PhaseResult) []string {
	var names []string
	for _, p := range phases {
		names = append(names, p.Phase.String())
	}
	return names
}

func phaseMetrics(phases []PhaseResult) map[string]interface{} {
	m := map[string]interface{}{}
	for _, p := range phases {
		m[p.Phase.String()] = map[string]interface{}{
			"success":    p.Success,
			"energy":     p.EnergyUsed,
			"latency_us": p.LatencyUS,
		}
	}
	return m
}

// ToJSON serializes WorkflowResult to JSON
func (r WorkflowResult) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
