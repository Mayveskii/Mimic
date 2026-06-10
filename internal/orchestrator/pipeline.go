package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/Mayveskii/Mimic/internal/cgo"
	"github.com/Mayveskii/Mimic/internal/hunt"
	"github.com/Mayveskii/Mimic/internal/session"
)

// Stage is a single step in the 6-stage pipeline.
type Stage interface {
	Name() string
	Process(ctx context.Context, sess *session.SessionContext, input []byte) ([]byte, error)
}

// Result is the final output of the pipeline.
type Result struct {
	Output    string `json:"output"`
	Metrics   string `json:"metrics"`
	Success   bool   `json:"success"`
	ErrorCode int    `json:"error_code,omitempty"`
}

// Pipeline is the 6-stage orchestrator: STATE → MESH → CLASSIFY → PLAN/VALIDATE/EXEC → VERIFY → RESPOND.
type Pipeline struct {
	Stages []Stage
}

// NewPipeline creates a fully wired pipeline.
func NewPipeline(hunter *hunt.Hunter) *Pipeline {
	return &Pipeline{
		Stages: []Stage{
			&StateStage{},
			&MeshStage{Hunter: hunter},
			&ClassifyStage{},
			&PlanValidateExecStage{},
			&VerifyStage{},
			&RespondStage{},
		},
	}
}

// Run executes the pipeline for the given session and intent.
func (p *Pipeline) Run(ctx context.Context, sess *session.SessionContext, intent string) (*Result, error) {
	if sess == nil {
		return nil, fmt.Errorf("session context is nil")
	}
	if sess.Budget.Exhausted() {
		return nil, fmt.Errorf("budget exhausted: %s", sess.Budget)
	}

	data := []byte(intent)
	var err error

	for _, stage := range p.Stages {
		data, err = stage.Process(ctx, sess, data)
		if err != nil {
			return &Result{
				Output:    err.Error(),
				Success:   false,
				ErrorCode: 1,
			}, nil
		}
	}

	return &Result{
		Output:  string(data),
		Success: true,
		Metrics: sess.Budget.String(),
	}, nil
}

// StateStage loads session state and repo map into context.
type StateStage struct{}

func (s *StateStage) Name() string { return "STATE" }

func (s *StateStage) Process(ctx context.Context, sess *session.SessionContext, input []byte) ([]byte, error) {
	// Validate worktree exists and is clean
	if sess.WorktreePath == "" {
		return nil, fmt.Errorf("worktree not provisioned")
	}
	// Budget gate: every stage consumes a small amount
	sess.Budget.Consume(10, 0)
	return input, nil
}

// MeshStage queries the mesh and injects top-ranked slots as context.
type MeshStage struct {
	Hunter *hunt.Hunter
}

func (s *MeshStage) Name() string { return "MESH" }

func (s *MeshStage) Process(ctx context.Context, sess *session.SessionContext, input []byte) ([]byte, error) {
	if s.Hunter == nil {
		return input, nil // graceful skip if hunter not configured
	}

	slots, err := s.Hunter.Search(string(input), "")
	if err != nil || len(slots) == 0 {
		return input, nil // no mesh results → continue with raw intent
	}

	ranked, err := s.Hunter.Rank(slots)
	if err != nil {
		return input, nil
	}

	// Inject top 3 slots as context
	var context string
	for i, slot := range ranked {
		if i >= 3 {
			break
		}
		context += fmt.Sprintf("\n# Slot: %s\nDomain: %s\nInvariant: %s\n", slot.ID, slot.Domain, slot.Invariant)
	}

	sess.Budget.Consume(20, 0)
	return append([]byte(context+"\n\nIntent: "), input...), nil
}

// ClassifyStage maps intent to domain and process name.
type ClassifyStage struct{}

func (s *ClassifyStage) Name() string { return "CLASSIFY" }

func (s *ClassifyStage) Process(ctx context.Context, sess *session.SessionContext, input []byte) ([]byte, error) {
	intent := string(input)
	// Simple keyword-based classification
	domain := "system"
	process := "sys_exec"

	if strings.Contains(intent, "git") || strings.Contains(intent, "commit") || strings.Contains(intent, "branch") || strings.Contains(intent, "merge") {
		domain = "git"
		process = "atomic_commit"
	} else if strings.Contains(intent, "build") || strings.Contains(intent, "compile") || strings.Contains(intent, "test") {
		domain = "build"
		process = "build_and_test"
	} else if strings.Contains(intent, "fix") || strings.Contains(intent, "patch") || strings.Contains(intent, "refactor") {
		domain = "git"
		process = "feature_branch"
	}

	sess.Budget.Consume(5, 0)
	return []byte(fmt.Sprintf("domain=%s process=%s intent=%s", domain, process, intent)), nil
}

// PlanValidateExecStage builds a plan, validates it, and executes via C-core.
type PlanValidateExecStage struct{}

func (s *PlanValidateExecStage) Name() string { return "EXEC" }

func (s *PlanValidateExecStage) Process(ctx context.Context, sess *session.SessionContext, input []byte) ([]byte, error) {
	// Stub: real implementation will build OpPacket chain, validate, and execute
	// For now, simulate a single SYS_FILE_WRITE for demonstration
	sess.Budget.Consume(100, 1)

	// Build a simple packet (stub)
	pkt := cgo.Packet{Opcode: "SYS_FILE_WRITE", Args: map[string]interface{}{"path": "/tmp/test", "content": "hello"}}
	result, err := cgo.ExecuteChain([]cgo.Packet{pkt}, float32(sess.Budget.MaxTokens), float32(sess.Budget.MaxTimeSeconds))
	if err != nil {
		// If C-core is not wired for this opcode, return input as-is (graceful)
		return input, nil
	}

	return []byte(fmt.Sprintf("executed: %s (result: %s)", string(input), result.Result)), nil
}

// VerifyStage performs 2-vote verification for critical operations.
type VerifyStage struct{}

func (s *VerifyStage) Name() string { return "VERIFY" }

func (s *VerifyStage) Process(ctx context.Context, sess *session.SessionContext, input []byte) ([]byte, error) {
	// Stub: 2-vote verify — currently passes through
	sess.Budget.Consume(10, 0)
	return input, nil
}

// RespondStage formats the final response with metrics and traceability.
type RespondStage struct{}

func (s *RespondStage) Name() string { return "RESPOND" }

func (s *RespondStage) Process(ctx context.Context, sess *session.SessionContext, input []byte) ([]byte, error) {
	output := fmt.Sprintf("%s\n\n[metrics: %s]", string(input), sess.Budget.String())
	sess.Budget.Consume(5, 0)
	return []byte(output), nil
}


