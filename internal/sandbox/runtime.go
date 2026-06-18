package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Mayveskii/Mimic/internal/cas"
)

// emptyTreeSHA is the canonical SHA-1 of an empty git tree object.
const emptyTreeSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// SandboxSpec describes a sandboxed execution environment and the commands to run.
type SandboxSpec struct {
	// BaseTreeSHA is the git tree SHA to materialize. An empty string is
	// treated as the empty tree.
	BaseTreeSHA string
	// WorkDir is the directory in which commands are executed. If empty, a
	// temporary directory is created and removed after the run.
	WorkDir string
	// Commands is a list of commands to execute sequentially. Each command is
	// represented as a list of arguments where the first argument is the
	// executable.
	Commands [][]string
	// Env is the environment for the executed commands. If nil, os.Environ()
	// is used.
	Env []string
}

// SandboxResult holds the outcome of a sandbox run.
type SandboxResult struct {
	// Output contains the combined stdout and stderr of all executed commands.
	Output string
	// NewTreeSHA is the git tree SHA captured after command execution.
	NewTreeSHA string
	// Proof contains mandatory proof artifacts collected after execution.
	Proof *Proof
	// Error is set when command execution fails.
	Error error
}

// Runtime executes commands inside an isolated git-backed workspace.
type Runtime struct {
	store *cas.Store
}

// NewRuntime creates a sandbox runtime backed by the given CAS store.
func NewRuntime(store *cas.Store) *Runtime {
	return &Runtime{store: store}
}

// Run materializes BaseTreeSHA, executes each command in WorkDir, captures the
// resulting tree SHA and proof artifacts, and returns the combined output.
// If a command fails, execution stops and the error is returned along with the
// output collected so far.
func (r *Runtime) Run(ctx context.Context, spec SandboxSpec) (*SandboxResult, error) {
	if r.store == nil {
		return nil, fmt.Errorf("runtime store is nil")
	}

	workDir := spec.WorkDir
	cleanup := false
	if workDir == "" {
		tmp, err := os.MkdirTemp("", "mimic-sandbox-*")
		if err != nil {
			return nil, fmt.Errorf("create sandbox dir: %w", err)
		}
		workDir = tmp
		cleanup = true
	}

	if cleanup {
		defer os.RemoveAll(workDir)
	}

	var err error
	workDir, err = filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve work dir: %w", err)
	}

	if err := initWorkDirRepo(ctx, workDir); err != nil {
		return nil, fmt.Errorf("init work dir repo: %w", err)
	}

	baseTree := spec.BaseTreeSHA
	if baseTree == "" {
		baseTree = emptyTreeSHA
	}

	if err := r.materializeTree(ctx, workDir, baseTree); err != nil {
		return nil, fmt.Errorf("materialize tree %s: %w", baseTree, err)
	}

	if err := initCommit(ctx, workDir); err != nil {
		return nil, fmt.Errorf("init commit: %w", err)
	}

	result := &SandboxResult{}
	env := spec.Env
	if env == nil {
		env = os.Environ()
	}

	var output bytes.Buffer
	for i, args := range spec.Commands {
		if len(args) == 0 {
			return nil, fmt.Errorf("command %d: empty args", i)
		}

		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Dir = workDir
		cmd.Env = env
		cmd.Stdout = &output
		cmd.Stderr = &output

		if err := cmd.Run(); err != nil {
			result.Output = output.String()
			result.Error = fmt.Errorf("command %d (%s) failed: %w", i, args[0], err)
			return result, result.Error
		}
	}
	result.Output = output.String()

	newTree, err := r.captureTree(ctx, workDir)
	if err != nil {
		return nil, fmt.Errorf("capture tree: %w", err)
	}
	result.NewTreeSHA = newTree

	proof, err := CollectProof(workDir)
	if err != nil {
		return nil, fmt.Errorf("collect proof: %w", err)
	}
	result.Proof = proof

	return result, nil
}

func initWorkDirRepo(ctx context.Context, workDir string) error {
	cmd := exec.CommandContext(ctx, "git", "init", workDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	_ = exec.CommandContext(ctx, "git", "-C", workDir, "config", "user.email", "mimic@localhost").Run()
	_ = exec.CommandContext(ctx, "git", "-C", workDir, "config", "user.name", "Mimic Sandbox").Run()
	return nil
}

func initCommit(ctx context.Context, workDir string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", workDir, "add", "-A")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	cmd = exec.CommandContext(ctx, "git", "-C", workDir, "commit", "-m", "mimic sandbox baseline", "--allow-empty")
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Mimic Sandbox",
		"GIT_AUTHOR_EMAIL=mimic@localhost",
		"GIT_COMMITTER_NAME=Mimic Sandbox",
		"GIT_COMMITTER_EMAIL=mimic@localhost",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (r *Runtime) materializeTree(ctx context.Context, workDir, treeSHA string) error {
	archiveCmd := exec.CommandContext(ctx, "git", "-C", r.store.RepoPath(), "archive", treeSHA)
	tarCmd := exec.CommandContext(ctx, "tar", "-x", "-C", workDir)

	pr, pw := io.Pipe()
	archiveCmd.Stdout = pw
	tarCmd.Stdin = pr

	var archiveErr, tarErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		tarErr = tarCmd.Run()
	}()

	archiveErr = archiveCmd.Run()
	_ = pw.Close()
	<-done

	if archiveErr != nil {
		return fmt.Errorf("git archive: %w", archiveErr)
	}
	if tarErr != nil {
		return fmt.Errorf("tar extract: %w", tarErr)
	}
	return nil
}

func (r *Runtime) captureTree(ctx context.Context, workDir string) (string, error) {
	addCmd := exec.CommandContext(ctx, "git", "-C", workDir, "add", "-A")
	if out, err := addCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add: %w\n%s", err, strings.TrimSpace(string(out)))
	}

	wtCmd := exec.CommandContext(ctx, "git", "-C", workDir, "write-tree")
	out, err := wtCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git write-tree: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
