package session

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mayveskii/Mimic/internal/config"
	"github.com/Mayveskii/Mimic/internal/model"
)

type mockCaller struct{}

func (m *mockCaller) Call(ctx context.Context, req model.ChatRequest) (*model.CallResult, error) {
	return &model.CallResult{Response: model.ChatResponse{}}, nil
}

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

	return repoPath
}

func TestManager_Init(t *testing.T) {
	repoPath := setupRepo(t)

	mgr := NewManager(repoPath)
	ctx, err := mgr.Init("qwen")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer mgr.Destroy(ctx)

	if ctx.ModelID != "qwen" {
		t.Fatalf("expected modelID=qwen, got %q", ctx.ModelID)
	}
	if ctx.WorktreePath == "" {
		t.Fatal("worktree path is empty")
	}
	if ctx.BaseSHA == "" {
		t.Fatal("base SHA is empty")
	}
	if ctx.Budget == nil {
		t.Fatal("budget is nil")
	}
	budget := ctx.Budget.(*Budget)
	if budget.MaxTokens != 100000 {
		t.Fatalf("expected max_tokens=100000, got %d", budget.MaxTokens)
	}
	if ctx.Config == nil {
		t.Fatal("config is nil")
	}

	// Verify worktree exists
	if _, err := os.Stat(ctx.WorktreePath); err != nil {
		t.Fatalf("worktree does not exist: %v", err)
	}
}

func TestManager_Init_DirtyRepo(t *testing.T) {
	repoPath := setupRepo(t)
	// Make repo dirty
	_ = os.WriteFile(filepath.Join(repoPath, "dirty.txt"), []byte("x"), 0644)

	mgr := NewManager(repoPath)
	_, err := mgr.Init("qwen")
	if err == nil {
		t.Fatal("expected error for dirty repo")
	}
	if !strings.Contains(err.Error(), "uncommitted changes") {
		t.Fatalf("expected dirty repo error, got: %v", err)
	}
}

func TestManager_Execute(t *testing.T) {
	repoPath := setupRepo(t)

	mgr := NewManager(repoPath)
	ctx, err := mgr.Init("qwen")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer mgr.Destroy(ctx)

	mgr = mgr.WithCaller(&mockCaller{})
	result, err := mgr.Execute(ctx, "fix race condition")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(result, "fix race condition") {
		t.Fatalf("expected result to contain intent, got: %s", result)
	}
	if ctx.Budget.(*Budget).UsedTokens == 0 {
		t.Fatal("expected budget to be consumed")
	}
}

func TestManager_Finalize(t *testing.T) {
	repoPath := setupRepo(t)

	mgr := NewManager(repoPath)
	ctx, err := mgr.Init("qwen")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer mgr.Destroy(ctx)

	// Modify worktree so proof has content
	_ = os.WriteFile(filepath.Join(ctx.WorktreePath, "new.go"), []byte("package main\n"), 0644)

	proof, err := mgr.Finalize(ctx)
	if err != nil {
		t.Fatalf("Finalize: %v", err)
	}
	if proof.GitStatus == "" {
		t.Fatal("expected non-empty git status")
	}
	if proof.GitLog == "" {
		t.Fatal("expected non-empty git log")
	}

	// Verify artifacts were written
	artifactsDir := filepath.Join(repoPath, ".mimic", "runs", "qwen")
	entries, err := os.ReadDir(artifactsDir)
	if err != nil {
		t.Fatalf("artifacts dir not created: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected artifacts to be written")
	}
}

func TestManager_Destroy(t *testing.T) {
	repoPath := setupRepo(t)

	mgr := NewManager(repoPath)
	ctx, err := mgr.Init("qwen")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}

	wtPath := ctx.WorktreePath
	if err := mgr.Destroy(ctx); err != nil {
		t.Fatalf("Destroy: %v", err)
	}
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Fatal("expected worktree to be removed")
	}
}

func TestBudget(t *testing.T) {
	b := NewBudget(1000, 60)
	if !b.Consume(100, 5) {
		t.Fatal("expected consume to succeed")
	}
	if b.UsedTokens != 100 {
		t.Fatalf("expected used tokens=100, got %d", b.UsedTokens)
	}

	// Exhaust budget
	if !b.Consume(900, 10) {
		t.Fatal("expected consume to succeed (exact limit)")
	}
	if b.Consume(1, 1) {
		t.Fatal("expected consume to fail (over budget)")
	}
	if !b.Exhausted() {
		t.Fatal("expected budget to be exhausted")
	}

	tokens, time := b.Remaining()
	if tokens != 0 {
		t.Fatalf("expected remaining tokens=0, got %d", tokens)
	}
	if time != 45 {
		t.Fatalf("expected remaining time=45, got %d", time)
	}
}

func TestLoadRepoConfig_Integration(t *testing.T) {
	repoPath := setupRepo(t)

	// Save custom config
	cfg := &config.RepoConfig{
		Repo:   "rtk",
		Models: config.ModelProfiles{Local: "qwen", Medium: "kimi", Top: "minimax"},
		Budget: config.BudgetConfig{MaxTokens: 50000, MaxIterations: 5, MaxTimeSeconds: 300},
	}
	if err := config.SaveRepoConfig(repoPath, cfg); err != nil {
		t.Fatalf("SaveRepoConfig: %v", err)
	}

	// Add .mimic to gitignore so repo stays clean
	_ = os.WriteFile(filepath.Join(repoPath, ".gitignore"), []byte(".mimic/\n"), 0644)
	_ = exec.Command("git", "-C", repoPath, "add", ".").Run()
	_ = exec.Command("git", "-C", repoPath, "commit", "-m", "add config").Run()

	mgr := NewManager(repoPath)
	ctx, err := mgr.Init("kimi")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer mgr.Destroy(ctx)

	budget := ctx.Budget.(*Budget)
	if budget.MaxTokens != 50000 {
		t.Fatalf("expected max_tokens=50000 from config, got %d", budget.MaxTokens)
	}
	if ctx.Config.Models.Medium != "kimi" {
		t.Fatalf("expected medium model=kimi from config, got %q", ctx.Config.Models.Medium)
	}
}
