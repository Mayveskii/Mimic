package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Mayveskii/Mimic/internal/config"
	"github.com/Mayveskii/Mimic/internal/core"
	"github.com/Mayveskii/Mimic/internal/model"
	"github.com/Mayveskii/Mimic/internal/orchestrator"
	"github.com/Mayveskii/Mimic/internal/sandbox"
)

// Manager handles session lifecycle: Init → Execute → Finalize → Destroy.
type Manager struct {
	Sandbox   *sandbox.WorktreeManager
	Caller    model.Caller
	Assembler *orchestrator.ContextAssembler
}

// NewManager creates a session manager for the given base repo.
func NewManager(baseRepo string) *Manager {
	return &Manager{
		Sandbox: sandbox.NewWorktreeManager(baseRepo),
	}
}

// WithCaller attaches an LLM caller to the manager.
func (m *Manager) WithCaller(caller model.Caller) *Manager {
	m.Caller = caller
	return m
}

// WithAssembler attaches a context assembler to the manager.
func (m *Manager) WithAssembler(assembler *orchestrator.ContextAssembler) *Manager {
	m.Assembler = assembler
	return m
}

// Init creates a new session: validates repo, provisions worktree, loads config, locks budget.
func (m *Manager) Init(modelID string) (*core.SessionContext, error) {
	// 1. Validate git reality
	sha, err := m.Sandbox.GetBaselineSHA()
	if err != nil {
		return nil, fmt.Errorf("validate git reality: %w", err)
	}

	dirty, err := m.Sandbox.IsDirty()
	if err != nil {
		return nil, fmt.Errorf("dirty check: %w", err)
	}
	if dirty {
		return nil, fmt.Errorf("base repo has uncommitted changes; commit or stash before running")
	}

	// 2. Load repo config
	cfg, err := config.LoadRepoConfig(m.Sandbox.BaseRepo)
	if err != nil {
		return nil, fmt.Errorf("load repo config: %w", err)
	}

	// 3. Provision worktree
	wtPath, err := m.Sandbox.Provision(modelID)
	if err != nil {
		return nil, fmt.Errorf("provision worktree: %w", err)
	}

	// 4. Lock budget
	budget := NewBudget(cfg.Budget.MaxTokens, cfg.Budget.MaxTimeSeconds)

	ctx := &core.SessionContext{
		ID:           fmt.Sprintf("%s-%s-%d", filepath.Base(m.Sandbox.BaseRepo), modelID, time.Now().Unix()),
		ModelID:      modelID,
		RepoPath:     m.Sandbox.BaseRepo,
		WorktreePath: wtPath,
		BaseSHA:      sha,
		Budget:       budget,
		Config:       cfg,
		StartTime:    time.Now(),
	}

	return ctx, nil
}

// Execute runs the model's intent in the session context.
func (m *Manager) Execute(ctx *core.SessionContext, intent string) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("session context is nil")
	}
	if ctx.Budget.Exhausted() {
		return "", fmt.Errorf("budget exhausted: %s", ctx.Budget)
	}
	if m.Caller == nil {
		return "", fmt.Errorf("no LLM caller configured")
	}

	pipeline := orchestrator.NewPipeline(nil, m.Caller, nil, m.Assembler)
	result, err := pipeline.Run(nil, ctx, intent)
	if err != nil {
		return "", fmt.Errorf("pipeline failed: %w", err)
	}
	if !result.Success {
		return "", fmt.Errorf("execution failed: %s", result.Output)
	}

	return result.Output, nil
}

// Finalize collects proof, writes artifacts, and extracts skills.
func (m *Manager) Finalize(ctx *core.SessionContext) (*sandbox.Proof, error) {
	if ctx == nil {
		return nil, fmt.Errorf("session context is nil")
	}

	// 1. Collect proof
	proof, err := sandbox.CollectProof(ctx.WorktreePath)
	if err != nil {
		return nil, fmt.Errorf("collect proof: %w", err)
	}

	// 2. Write artifacts
	artifactsDir := filepath.Join(ctx.RepoPath, ".mimic", "runs", ctx.ModelID, fmt.Sprintf("%d", ctx.StartTime.Unix()))
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir artifacts: %w", err)
	}

	// Write proof files
	_ = os.WriteFile(filepath.Join(artifactsDir, "git_status.txt"), []byte(proof.GitStatus), 0644)
	_ = os.WriteFile(filepath.Join(artifactsDir, "git_log.txt"), []byte(proof.GitLog), 0644)
	_ = os.WriteFile(filepath.Join(artifactsDir, "ls_la.txt"), []byte(proof.LsLa), 0644)
	if proof.GitDiff != "" {
		_ = os.WriteFile(filepath.Join(artifactsDir, "git_diff.txt"), []byte(proof.GitDiff), 0644)
	}

	// TODO: Normalize output → []Finding
	// TODO: Extract skills → mesh slot

	return proof, nil
}

// Destroy cleans up the session: destroys worktree.
func (m *Manager) Destroy(ctx *core.SessionContext) error {
	if ctx == nil {
		return fmt.Errorf("session context is nil")
	}
	if err := m.Sandbox.Destroy(ctx.ModelID); err != nil {
		return fmt.Errorf("destroy worktree: %w", err)
	}
	return nil
}
