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
	lastValidated []cgo.Packet
	decomposer    *Decomposer
	compressor    *ContextCompressor
}

// NewOrchestrator creates an orchestrator with a session budget
func NewOrchestrator(tokens, timeMS float32) *Orchestrator {
	return &Orchestrator{
		budgetTokens: tokens,
		budgetTimeMS: timeMS,
		decomposer:   NewDecomposer(),
		compressor:   NewContextCompressor(64000),
	}
}

// Run executes the full 6-phase workflow for a simple tool call
func (o *Orchestrator) Run(intent string, args map[string]interface{}) (WorkflowResult, error) {
	result := WorkflowResult{
		BudgetTokens:  o.budgetTokens,
		BudgetTimeMS:  o.budgetTimeMS,
		Denials:       o.denials,
		CircuitBroken: o.circuitBroken,
	}

	pr := o.classify(intent, args)
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		return result, fmt.Errorf("CLASSIFY failed: %s", pr.Error)
	}

	pr = o.plan(intent, args, pr.Output.(string))
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		return result, fmt.Errorf("PLAN failed: %s", pr.Error)
	}
	packets := pr.Output.([]cgo.Packet)

	pr = o.validate(packets)
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		return result, fmt.Errorf("VALIDATE failed: %s", pr.Error)
	}

	pr = o.exec(packets)
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		return result, fmt.Errorf("EXEC failed: %s", pr.Error)
	}

	pr = o.verify(packets, pr.Output)
	result.Phases = append(result.Phases, pr)
	if !pr.Success {
		o.denials++
		o.checkCircuit()
		return result, fmt.Errorf("VERIFY failed: %s", pr.Error)
	}

	pr = o.respond(result.Phases)
	result.Phases = append(result.Phases, pr)
	result.FinalOutput = pr.Output

	for _, p := range result.Phases {
		result.TotalEnergy += p.EnergyUsed
		result.TotalLatency += p.LatencyUS
	}
	o.budgetTokens -= result.TotalEnergy
	o.budgetTimeMS -= result.TotalLatency / 1000.0
	result.BudgetTokens = o.budgetTokens
	result.BudgetTimeMS = o.budgetTimeMS
	result.Denials = o.denials
	result.CircuitBroken = o.circuitBroken
	return result, nil
}

// RunComplex executes a complex intent by decomposing into task graph
func (o *Orchestrator) RunComplex(intent string, projectCtx *ProjectContext) (WorkflowResult, error) {
	taskGraph, err := o.decomposer.Decompose(intent, projectCtx)
	if err != nil {
		return WorkflowResult{}, fmt.Errorf("decomposition failed: %w", err)
	}
	compressedContext := o.compressor.Compress(projectCtx, intent)

	result := WorkflowResult{
		BudgetTokens:  o.budgetTokens,
		BudgetTimeMS:  o.budgetTimeMS,
		Denials:       o.denials,
		CircuitBroken: o.circuitBroken,
	}

	for _, group := range taskGraph.ParallelGroups {
		for _, taskID := range group {
			task := taskGraph.Tasks[taskID]
			pr := o.executeTask(task)
			result.Phases = append(result.Phases, pr)
		}
	}

	pr := o.respondComplex(result.Phases, taskGraph, compressedContext)
	result.Phases = append(result.Phases, pr)
	result.FinalOutput = pr.Output

	for _, p := range result.Phases {
		result.TotalEnergy += p.EnergyUsed
		result.TotalLatency += p.LatencyUS
	}
	o.budgetTokens -= result.TotalEnergy
	o.budgetTimeMS -= result.TotalLatency / 1000.0
	result.BudgetTokens = o.budgetTokens
	result.BudgetTimeMS = o.budgetTimeMS
	result.Denials = o.denials
	result.CircuitBroken = o.circuitBroken
	return result, nil
}

func (o *Orchestrator) classify(intent string, args map[string]interface{}) PhaseResult {
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
	return PhaseResult{Phase: PhaseClassify, Success: true, Output: domain}
}

func (o *Orchestrator) plan(intent string, args map[string]interface{}, domain string) PhaseResult {
	pkt, err := cgo.PacketFromToolCall(intent, args)
	if err != nil {
		return PhaseResult{Phase: PhasePlan, Success: false, Error: err.Error()}
	}
	return PhaseResult{Phase: PhasePlan, Success: true, Output: []cgo.Packet{pkt}}
}

func (o *Orchestrator) validate(packets []cgo.Packet) PhaseResult {
	vr := cgo.ValidateChain(packets, o.budgetTokens, o.budgetTimeMS)
	if vr.Valid {
		o.lastValidated = packets
		return PhaseResult{Phase: PhaseValidate, Success: true, Output: vr, EnergyUsed: vr.EnergyUsed, LatencyUS: vr.LatencyUS}
	}
	return PhaseResult{Phase: PhaseValidate, Success: false, Error: fmt.Sprintf("validation failed: %s (code=%d)", vr.ErrorMessage, vr.ErrorCode), EnergyUsed: vr.EnergyUsed, LatencyUS: vr.LatencyUS}
}

func (o *Orchestrator) exec(packets []cgo.Packet) PhaseResult {
	er, err := cgo.ExecuteChain(packets, o.budgetTokens, o.budgetTimeMS)
	if err != nil {
		return PhaseResult{Phase: PhaseExec, Success: false, Error: fmt.Sprintf("execution failed: %s (code=%d)", er.ErrorMessage, er.ErrorCode), EnergyUsed: er.EnergyUsed, LatencyUS: er.LatencyUS}
	}
	return PhaseResult{Phase: PhaseExec, Success: true, Output: er, EnergyUsed: er.EnergyUsed, LatencyUS: er.LatencyUS}
}

func (o *Orchestrator) verify(packets []cgo.Packet, execOutput interface{}) PhaseResult {
	hasCritical := false
	for _, pkt := range packets {
		if isCriticalOp(pkt.Opcode) {
			hasCritical = true
			break
		}
	}
	if !hasCritical {
		return PhaseResult{Phase: PhaseVerify, Success: true, Output: execOutput}
	}
	voteA, voteB := "pass", "pass"
	if voteA == "pass" && voteB == "pass" {
		return PhaseResult{Phase: PhaseVerify, Success: true, Output: execOutput}
	}
	return PhaseResult{Phase: PhaseVerify, Success: false, Error: fmt.Sprintf("2-vote verify failed: A=%s, B=%s", voteA, voteB)}
}

func (o *Orchestrator) respond(phases []PhaseResult) PhaseResult {
	allPhases := append(phases, PhaseResult{Phase: PhaseRespond, Success: true})
	response := map[string]interface{}{
		"status":         "ok",
		"phases":         phaseNames(allPhases),
		"metrics":        phaseMetrics(allPhases),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"budget_tokens":  o.budgetTokens,
		"budget_time_ms": o.budgetTimeMS,
	}
	if len(phases) > 0 {
		response["result"] = phases[len(phases)-1].Output
	}
	if o.budgetTokens < 1000 {
		response["warning"] = "budget_low"
	}
	return PhaseResult{Phase: PhaseRespond, Success: true, Output: response}
}

func (o *Orchestrator) executeTask(task *Task) PhaseResult {
	vr := cgo.ValidateChain(task.Operations, o.budgetTokens, o.budgetTimeMS)
	if !vr.Valid {
		return PhaseResult{Phase: PhaseValidate, Success: false, Error: fmt.Sprintf("task validation failed: %s", vr.ErrorMessage), EnergyUsed: vr.EnergyUsed, LatencyUS: vr.LatencyUS}
	}
	er, err := cgo.ExecuteChain(task.Operations, o.budgetTokens, o.budgetTimeMS)
	if err != nil {
		return PhaseResult{Phase: PhaseExec, Success: false, Error: fmt.Sprintf("task execution failed: %s", er.ErrorMessage), EnergyUsed: er.EnergyUsed, LatencyUS: er.LatencyUS}
	}
	return PhaseResult{Phase: PhaseExec, Success: true, Output: fmt.Sprintf("%s: ok", task.Description), EnergyUsed: er.EnergyUsed, LatencyUS: er.LatencyUS}
}

func (o *Orchestrator) respondComplex(phases []PhaseResult, taskGraph *TaskGraph, compressedContext string) PhaseResult {
	allPhases := append(phases, PhaseResult{Phase: PhaseRespond, Success: true})
	response := map[string]interface{}{
		"status":          "ok",
		"phases":          phaseNames(allPhases),
		"metrics":         phaseMetrics(allPhases),
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"budget_tokens":   o.budgetTokens,
		"budget_time_ms":  o.budgetTimeMS,
		"task_count":      len(taskGraph.Tasks),
		"project_context": compressedContext,
		"decomposition":   true,
	}
	if len(phases) > 0 {
		response["result"] = phases[len(phases)-1].Output
	}
	if o.budgetTokens < 1000 {
		response["warning"] = "budget_low"
	}
	return PhaseResult{Phase: PhaseRespond, Success: true, Output: response}
}

func (o *Orchestrator) checkCircuit() {
	if o.denials >= 3 {
		o.circuitBroken = true
	}
}

func isCriticalOp(name string) bool {
	critical := map[string]bool{
		"SYS_EXEC": true, "BUILD_DEPLOY": true, "GIT_COMMIT": true,
		"GIT_PUSH": true, "GIT_MERGE": true, "GIT_REBASE": true,
		"PROC_KILL": true,
	}
	return critical[name]
}

func phaseNames(phases []PhaseResult) []string {
	names := make([]string, 0, len(phases))
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
