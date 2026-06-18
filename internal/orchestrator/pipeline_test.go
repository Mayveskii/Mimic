package orchestrator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mayveskii/Mimic/internal/config"
	"github.com/Mayveskii/Mimic/internal/core"
	"github.com/Mayveskii/Mimic/internal/hunt"
	"github.com/Mayveskii/Mimic/internal/mesh"
	"github.com/Mayveskii/Mimic/internal/sandbox"
)

type testBudget struct {
	maxTokens       int
	maxTimeSeconds  int
	usedTokens      int
	usedTimeSeconds int
}

func (b *testBudget) Consume(tokens, seconds int) bool {
	if b.usedTokens+tokens > b.maxTokens {
		return false
	}
	if b.usedTimeSeconds+seconds > b.maxTimeSeconds {
		return false
	}
	b.usedTokens += tokens
	b.usedTimeSeconds += seconds
	return true
}

func (b *testBudget) Exhausted() bool {
	return b.usedTokens >= b.maxTokens || b.usedTimeSeconds >= b.maxTimeSeconds
}

func (b *testBudget) String() string {
	return fmt.Sprintf("tokens=%d/%d time=%ds/%ds", b.usedTokens, b.maxTokens, b.usedTimeSeconds, b.maxTimeSeconds)
}

func (b *testBudget) TokenBudget() int { return b.maxTokens }
func (b *testBudget) TimeBudget() int  { return b.maxTimeSeconds }

func setupRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "rtk")
	_ = os.MkdirAll(repoPath, 0755)
	_ = exec.Command("git", "init", repoPath).Run()
	_ = exec.Command("git", "-C", repoPath, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", repoPath, "config", "user.name", "Test").Run()
	_ = os.WriteFile(filepath.Join(repoPath, "main.go"), []byte("package main\n"), 0644)
	_ = exec.Command("git", "-C", repoPath, "add", ".").Run()
	_ = exec.Command("git", "-C", repoPath, "commit", "-m", "init").Run()
	_ = os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte(".mimic/\n"), 0644)
	_ = exec.Command("git", "-C", repoPath, "add", ".").Run()
	_ = exec.Command("git", "-C", repoPath, "commit", "-m", "gitignore").Run()
	return repoPath
}

func TestPipeline_Run_GitIntent(t *testing.T) {
	repoPath := setupRepo(t)

	// Setup mesh with a git slot
	meshDir := filepath.Join(repoPath, ".mimic", "slots")
	_ = os.MkdirAll(filepath.Join(meshDir, "git"), 0755)
	slot := &mesh.TextSlot{
		ID:        "git-1",
		Domain:    "git",
		Invariant: "never commit without reviewing diff first",
		Context:   "from embryo",
		Metadata: map[string]string{
			"survival_index": "0.90",
			"z_density":      "0.85",
		},
	}
	_ = slot.SaveToFile(meshDir)

	hunter := hunt.NewHunter(meshDir)
	pipeline := NewPipeline(hunter, nil, nil, nil)

	// Setup session
	wm := sandbox.NewWorktreeManager(repoPath)
	wtPath, err := wm.Provision("qwen")
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	defer wm.Destroy("qwen")

	sess := &core.SessionContext{
		ID:           "test-1",
		ModelID:      "qwen",
		RepoPath:     repoPath,
		WorktreePath: wtPath,
		Budget:       &testBudget{maxTokens: 100000, maxTimeSeconds: 600},
		Config:       config.DefaultRepoConfig(),
	}

	result, err := pipeline.Run(context.Background(), sess, "commit these files")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got failure: %s", result.Output)
	}
	if !strings.Contains(result.Output, "domain=git") {
		t.Fatalf("expected classification to git domain, got: %s", result.Output)
	}
	if !strings.Contains(result.Output, "metrics:") {
		t.Fatalf("expected metrics in output, got: %s", result.Output)
	}
}

func TestPipeline_Run_BuildIntent(t *testing.T) {
	repoPath := setupRepo(t)
	hunter := hunt.NewHunter("")
	pipeline := NewPipeline(hunter, nil, nil, nil)

	wm := sandbox.NewWorktreeManager(repoPath)
	wtPath, err := wm.Provision("kimi")
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	defer wm.Destroy("kimi")

	sess := &core.SessionContext{
		ID:           "test-2",
		ModelID:      "kimi",
		RepoPath:     repoPath,
		WorktreePath: wtPath,
		Budget:       &testBudget{maxTokens: 100000, maxTimeSeconds: 600},
		Config:       config.DefaultRepoConfig(),
	}

	result, err := pipeline.Run(context.Background(), sess, "build the project")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got failure: %s", result.Output)
	}
	if !strings.Contains(result.Output, "domain=build") {
		t.Fatalf("expected classification to build domain, got: %s", result.Output)
	}
}

func TestPipeline_BudgetExhausted(t *testing.T) {
	repoPath := setupRepo(t)
	hunter := hunt.NewHunter("")
	pipeline := NewPipeline(hunter, nil, nil, nil)

	wm := sandbox.NewWorktreeManager(repoPath)
	wtPath, err := wm.Provision("minimax")
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	defer wm.Destroy("minimax")

	// Exhausted budget
	sess := &core.SessionContext{
		ID:           "test-3",
		ModelID:      "minimax",
		RepoPath:     repoPath,
		WorktreePath: wtPath,
		Budget:       &testBudget{maxTokens: 1, maxTimeSeconds: 1, usedTokens: 1, usedTimeSeconds: 1},
		Config:       config.DefaultRepoConfig(),
	}

	result, err := pipeline.Run(context.Background(), sess, "any intent")
	if err == nil {
		t.Fatal("expected error for exhausted budget")
	}
	if !strings.Contains(err.Error(), "budget exhausted") {
		t.Fatalf("expected budget exhausted error, got: %v", err)
	}
	_ = result
}

func TestPipeline_Classify_KeywordCases(t *testing.T) {
	tests := []struct {
		intent     string
		wantDomain string
	}{
		{"commit these files", "git"},
		{"create a branch", "git"},
		{"merge feature into main", "git"},
		{"build the project", "build"},
		{"run tests", "build"},
		{"compile all", "build"},
		{"fix race condition", "git"},
		{"patch the bug", "git"},
		{"check status", "system"},
	}

	for _, tt := range tests {
		t.Run(tt.intent, func(t *testing.T) {
			stage := &ClassifyStage{}
			out, err := stage.Process(context.Background(), &core.SessionContext{Budget: &testBudget{maxTokens: 100000, maxTimeSeconds: 600}}, []byte(tt.intent))
			if err != nil {
				t.Fatalf("Process: %v", err)
			}
			if !strings.Contains(string(out), "domain="+tt.wantDomain) {
				t.Fatalf("expected domain=%s in %s", tt.wantDomain, string(out))
			}
		})
	}
}
