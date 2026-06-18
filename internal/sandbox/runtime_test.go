package sandbox

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mayveskii/Mimic/internal/cas"
)

func setupCASStore(t *testing.T) (*cas.Store, string) {
	t.Helper()
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "store-repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("mkdir store repo: %v", err)
	}

	cmd := exec.Command("git", "init", repoPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	_ = exec.Command("git", "-C", repoPath, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", repoPath, "config", "user.name", "Test").Run()

	helloPath := filepath.Join(repoPath, "hello.txt")
	if err := os.WriteFile(helloPath, []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write hello.txt: %v", err)
	}
	_ = exec.Command("git", "-C", repoPath, "add", ".").Run()
	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "init")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	store, err := cas.NewStore(repoPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return store, repoPath
}

func treeSHA(t *testing.T, repoPath string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD^{tree}").Output()
	if err != nil {
		t.Fatalf("rev-parse tree: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestRuntime_RunEmptyTreeCreatesFile(t *testing.T) {
	store, _ := setupCASStore(t)
	rt := NewRuntime(store)
	workDir := filepath.Join(t.TempDir(), "sandbox")

	res, err := rt.Run(context.Background(), SandboxSpec{
		WorkDir:  workDir,
		Commands: [][]string{{"sh", "-c", "echo sandbox > out.txt"}},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.NewTreeSHA == "" {
		t.Fatal("NewTreeSHA is empty")
	}
	if res.Proof == nil {
		t.Fatal("Proof is nil")
	}

	content, err := os.ReadFile(filepath.Join(workDir, "out.txt"))
	if err != nil {
		t.Fatalf("read out.txt: %v", err)
	}
	if strings.TrimSpace(string(content)) != "sandbox" {
		t.Fatalf("unexpected out.txt content: %q", string(content))
	}
}

func TestRuntime_RunModifiesExistingFile(t *testing.T) {
	store, repoPath := setupCASStore(t)
	baseTree := treeSHA(t, repoPath)
	rt := NewRuntime(store)
	workDir := filepath.Join(t.TempDir(), "sandbox")

	res, err := rt.Run(context.Background(), SandboxSpec{
		BaseTreeSHA: baseTree,
		WorkDir:     workDir,
		Commands:    [][]string{{"sh", "-c", "echo modified > hello.txt"}},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.NewTreeSHA == "" {
		t.Fatal("NewTreeSHA is empty")
	}
	if res.NewTreeSHA == baseTree {
		t.Fatalf("NewTreeSHA should differ from BaseTreeSHA")
	}

	content, err := os.ReadFile(filepath.Join(workDir, "hello.txt"))
	if err != nil {
		t.Fatalf("read hello.txt: %v", err)
	}
	if strings.TrimSpace(string(content)) != "modified" {
		t.Fatalf("unexpected hello.txt content: %q", string(content))
	}
}

func TestRuntime_RunCommandFailure(t *testing.T) {
	store, _ := setupCASStore(t)
	rt := NewRuntime(store)
	workDir := filepath.Join(t.TempDir(), "sandbox")

	res, err := rt.Run(context.Background(), SandboxSpec{
		WorkDir: workDir,
		Commands: [][]string{
			{"sh", "-c", "echo before"},
			{"false"},
			{"sh", "-c", "echo after"},
		},
	})
	if err == nil {
		t.Fatal("expected error for failed command")
	}
	if res == nil {
		t.Fatal("expected result even on failure")
	}
	if !strings.Contains(res.Output, "before") {
		t.Fatalf("expected 'before' in output, got %q", res.Output)
	}
	if strings.Contains(res.Output, "after") {
		t.Fatal("expected 'after' not to run")
	}
	if res.Error == nil {
		t.Fatal("expected result.Error to be set")
	}
}
