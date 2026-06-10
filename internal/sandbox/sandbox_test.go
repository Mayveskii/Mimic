package sandbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	baseRepo := filepath.Join(tmpDir, "base-repo")
	if err := os.MkdirAll(baseRepo, 0755); err != nil {
		t.Fatalf("mkdir base repo: %v", err)
	}

	cmd := exec.Command("git", "init", baseRepo)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	_ = exec.Command("git", "-C", baseRepo, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", baseRepo, "config", "user.name", "Test").Run()

	testFile := filepath.Join(baseRepo, "hello.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_ = exec.Command("git", "-C", baseRepo, "add", ".").Run()
	cmd = exec.Command("git", "-C", baseRepo, "commit", "-m", "init")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	return baseRepo
}

func TestWorktreeManager_ProvisionDestroy(t *testing.T) {
	baseRepo := setupRepo(t)
	wm := NewWorktreeManager(baseRepo)

	// Get baseline SHA
	sha, err := wm.GetBaselineSHA()
	if err != nil {
		t.Fatalf("GetBaselineSHA: %v", err)
	}
	if sha == "" {
		t.Fatal("baseline SHA is empty")
	}

	// Check dirty
	dirty, err := wm.IsDirty()
	if err != nil {
		t.Fatalf("IsDirty: %v", err)
	}
	if dirty {
		t.Fatal("expected clean repo")
	}

	// Provision worktree
	wtPath, err := wm.Provision("test-model")
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if wtPath == "" {
		t.Fatal("worktree path is empty")
	}
	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("worktree directory does not exist: %v", err)
	}

	// Verify worktree is detached
	cmd := exec.Command("git", "-C", wtPath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	if strings.TrimSpace(string(out)) != "HEAD" {
		t.Fatalf("expected detached HEAD, got %q", string(out))
	}

	// Verify worktree has the file
	wtFile := filepath.Join(wtPath, "hello.txt")
	if _, err := os.Stat(wtFile); err != nil {
		t.Fatalf("worktree missing hello.txt: %v", err)
	}

	// Destroy worktree
	if err := wm.Destroy("test-model"); err != nil {
		t.Fatalf("Destroy: %v", err)
	}
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Fatalf("worktree directory still exists after destroy")
	}
}

func TestCollectProof(t *testing.T) {
	// Create a temporary base repository
	tmpDir := t.TempDir()
	baseRepo := filepath.Join(tmpDir, "base-repo")
	_ = os.MkdirAll(baseRepo, 0755)

	cmd := exec.Command("git", "init", baseRepo)
	_ = cmd.Run()
	_ = exec.Command("git", "-C", baseRepo, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", baseRepo, "config", "user.name", "Test").Run()

	_ = os.WriteFile(filepath.Join(baseRepo, "hello.txt"), []byte("hello"), 0644)
	_ = exec.Command("git", "-C", baseRepo, "add", ".").Run()
	_ = exec.Command("git", "-C", baseRepo, "commit", "-m", "init").Run()

	wm := NewWorktreeManager(baseRepo)
	wtPath, err := wm.Provision("test-proof")
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	defer wm.Destroy("test-proof")

	// Modify a file in worktree
	_ = os.WriteFile(filepath.Join(wtPath, "hello.txt"), []byte("modified"), 0644)

	// Collect proof
	proof, err := CollectProof(wtPath)
	if err != nil {
		t.Fatalf("CollectProof: %v", err)
	}
	if proof.GitStatus == "" {
		t.Fatal("expected non-empty git status")
	}
	if proof.GitLog == "" {
		t.Fatal("expected non-empty git log")
	}
	if proof.LsLa == "" {
		t.Fatal("expected non-empty ls -la")
	}
	if proof.GitDiff == "" {
		t.Fatal("expected non-empty git diff (we modified a file)")
	}
}

func TestRollback(t *testing.T) {
	// Create a temporary base repository
	tmpDir := t.TempDir()
	baseRepo := filepath.Join(tmpDir, "base-repo")
	_ = os.MkdirAll(baseRepo, 0755)

	cmd := exec.Command("git", "init", baseRepo)
	_ = cmd.Run()
	_ = exec.Command("git", "-C", baseRepo, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", baseRepo, "config", "user.name", "Test").Run()

	_ = os.WriteFile(filepath.Join(baseRepo, "hello.txt"), []byte("hello"), 0644)
	_ = exec.Command("git", "-C", baseRepo, "add", ".").Run()
	_ = exec.Command("git", "-C", baseRepo, "commit", "-m", "init").Run()

	wm := NewWorktreeManager(baseRepo)
	wtPath, err := wm.Provision("test-rollback")
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	defer wm.Destroy("test-rollback")

	// Modify file and create untracked file
	_ = os.WriteFile(filepath.Join(wtPath, "hello.txt"), []byte("modified"), 0644)
	_ = os.WriteFile(filepath.Join(wtPath, "untracked.txt"), []byte("new"), 0644)

	// Rollback
	if err := Rollback(wtPath); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	// Verify file is back to original
	content, err := os.ReadFile(filepath.Join(wtPath, "hello.txt"))
	if err != nil {
		t.Fatalf("read hello.txt after rollback: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("expected 'hello' after rollback, got %q", string(content))
	}

	// Verify untracked file is gone
	if _, err := os.Stat(filepath.Join(wtPath, "untracked.txt")); !os.IsNotExist(err) {
		t.Fatal("expected untracked.txt to be removed after rollback")
	}

	// Verify clean status
	cmd = exec.Command("git", "-C", wtPath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status after rollback: %v", err)
	}
	if len(out) > 0 {
		t.Fatalf("expected clean tree after rollback, got:\n%s", string(out))
	}
}
