package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Mayveskii/Mimic/internal/cas"
	"github.com/Mayveskii/Mimic/internal/mesh"
	"github.com/Mayveskii/Mimic/internal/sandbox"
)

func main() {
	tmpDir, err := os.MkdirTemp("", "mimic-cas-demo-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	repoPath := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir repo: %v\n", err)
		os.Exit(1)
	}
	if err := initGitRepo(repoPath); err != nil {
		fmt.Fprintf(os.Stderr, "git init: %v\n", err)
		os.Exit(1)
	}

	store, err := cas.NewStore(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "new store: %v\n", err)
		os.Exit(1)
	}
	assembler := cas.NewAssembler(store)

	// Register a proven pattern for a Go main package.
	pattern := &mesh.TextSlot{
		ID:        "go-main-skeleton",
		Domain:    "go",
		Invariant: "package main with a run function",
		Context:   "Minimal Go entry point",
		Actions:   []string{"package main", "", "import \"fmt\"", "", "func main() {", "\tfmt.Println(\"${message}\")", "}", ""},
	}
	patternSHA, err := store.RegisterPattern(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "register pattern: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Registered pattern: %s\n", patternSHA)

	// Assemble: write a README and apply the pattern.
	steps := []cas.AssemblyStep{
		cas.WriteFileStep{Path: "README.md", Content: []byte("# CAS Demo\n")},
		cas.ApplyPatternStep{PatternSHA: patternSHA, TargetPath: "main.go", Params: map[string]string{"message": "hello from SHA-based sandbox"}},
	}
	outputTreeSHA, err := assembler.Assemble(context.Background(), cas.EmptyTreeSHA, steps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "assemble: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Output tree SHA: %s\n", outputTreeSHA)

	// Commit the transition.
	session := sandbox.NewSessionRef(store)
	uuid := "demo-session"
	if err := session.Create(uuid, outputTreeSHA); err != nil {
		fmt.Fprintf(os.Stderr, "create session ref: %v\n", err)
		os.Exit(1)
	}
	commitSHA, err := session.Advance(uuid, outputTreeSHA, "cas demo: assemble main.go and README")
	if err != nil {
		fmt.Fprintf(os.Stderr, "advance session: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Commit SHA: %s\n", commitSHA)

	// Run sandbox: materialize tree and execute a command.
	runtime := sandbox.NewRuntime(store)
	result, err := runtime.Run(context.Background(), sandbox.SandboxSpec{
		BaseTreeSHA: outputTreeSHA,
		Commands:    [][]string{{"cat", "main.go"}, {"cat", "README.md"}},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "runtime: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Runtime output:\n%s\n", result.Output)
	fmt.Printf("New tree SHA after runtime: %s\n", result.NewTreeSHA)
	fmt.Printf("Proof collected: git status (%d bytes), git diff (%d bytes), git log (%d bytes)\n",
		len(result.Proof.GitStatus), len(result.Proof.GitDiff), len(result.Proof.GitLog))

	fmt.Println("\nCAS demo complete. Every state is a SHA.")
}

func initGitRepo(path string) error {
	cmd := exec.Command("git", "init", path)
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "-C", path, "config", "user.email", "demo@mimic.dev")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "-C", path, "config", "user.name", "Mimic Demo")
	return cmd.Run()
}
