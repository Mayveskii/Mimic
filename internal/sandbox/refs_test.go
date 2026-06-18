package sandbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Mayveskii/Mimic/internal/cas"
)

func setupSessionStore(t *testing.T) (*cas.Store, string) {
	t.Helper()
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "session-repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	cmd := exec.Command("git", "init", repoPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	_ = exec.Command("git", "-C", repoPath, "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "-C", repoPath, "config", "user.name", "Test").Run()

	testFile := filepath.Join(repoPath, "hello.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
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

func makeTree(t *testing.T, repoPath string, files map[string]string) string {
	t.Helper()
	for path, data := range files {
		fullPath := filepath.Join(repoPath, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(data), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
	_ = exec.Command("git", "-C", repoPath, "add", "-A").Run()
	out, err := exec.Command("git", "-C", repoPath, "write-tree").Output()
	if err != nil {
		t.Fatalf("write-tree: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestSessionRef_CreateAndAdvance(t *testing.T) {
	store, repoPath := setupSessionStore(t)
	ref := NewSessionRef(store)

	baseTree := makeTree(t, repoPath, map[string]string{"base.txt": "base"})
	tree1 := makeTree(t, repoPath, map[string]string{"a.txt": "a"})
	tree2 := makeTree(t, repoPath, map[string]string{"b.txt": "b"})

	if err := ref.Create("session-1", baseTree); err != nil {
		t.Fatalf("Create: %v", err)
	}

	commit1, err := ref.Advance("session-1", tree1, "first")
	if err != nil {
		t.Fatalf("Advance first: %v", err)
	}
	if commit1 == "" {
		t.Fatal("first commit SHA is empty")
	}

	commit2, err := ref.Advance("session-1", tree2, "second")
	if err != nil {
		t.Fatalf("Advance second: %v", err)
	}
	if commit2 == "" {
		t.Fatal("second commit SHA is empty")
	}
	if commit1 == commit2 {
		t.Fatal("commit SHAs should differ")
	}

	parent, err := exec.Command("git", "-C", repoPath, "rev-parse", commit2+"^").Output()
	if err != nil {
		t.Fatalf("resolve commit2 parent: %v", err)
	}
	if strings.TrimSpace(string(parent)) != commit1 {
		t.Fatalf("expected commit2 parent %s, got %s", commit1, strings.TrimSpace(string(parent)))
	}
}

func TestSessionRef_Rewind(t *testing.T) {
	store, repoPath := setupSessionStore(t)
	ref := NewSessionRef(store)

	baseTree := makeTree(t, repoPath, map[string]string{"base.txt": "base"})
	tree1 := makeTree(t, repoPath, map[string]string{"a.txt": "a"})
	tree2 := makeTree(t, repoPath, map[string]string{"b.txt": "b"})

	if err := ref.Create("session-2", baseTree); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := ref.Advance("session-2", tree1, "first"); err != nil {
		t.Fatalf("Advance: %v", err)
	}
	if err := ref.Rewind("session-2", tree2); err != nil {
		t.Fatalf("Rewind: %v", err)
	}

	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "refs/sessions/session-2").Output()
	if err != nil {
		t.Fatalf("rev-parse ref: %v", err)
	}
	if strings.TrimSpace(string(out)) != tree2 {
		t.Fatalf("expected ref to point to %s, got %s", tree2, strings.TrimSpace(string(out)))
	}
}

func TestSessionRef_History(t *testing.T) {
	store, repoPath := setupSessionStore(t)
	ref := NewSessionRef(store)

	baseTree := makeTree(t, repoPath, map[string]string{"base.txt": "base"})
	tree1 := makeTree(t, repoPath, map[string]string{"a.txt": "a"})
	tree2 := makeTree(t, repoPath, map[string]string{"b.txt": "b"})

	if err := ref.Create("session-3", baseTree); err != nil {
		t.Fatalf("Create: %v", err)
	}
	commit1, err := ref.Advance("session-3", tree1, "first")
	if err != nil {
		t.Fatalf("Advance first: %v", err)
	}
	commit2, err := ref.Advance("session-3", tree2, "second")
	if err != nil {
		t.Fatalf("Advance second: %v", err)
	}

	history, err := ref.History("session-3")
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(history))
	}
	if history[0] != commit2 {
		t.Fatalf("expected newest commit %s, got %s", commit2, history[0])
	}
	if history[1] != commit1 {
		t.Fatalf("expected older commit %s, got %s", commit1, history[1])
	}
}
