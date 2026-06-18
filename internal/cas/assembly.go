package cas

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

// AssemblyStep is a single operation in an assembly plan.
type AssemblyStep interface {
	Kind() string
}

// ApplyPatternStep writes a stored pattern's content into the workspace.
type ApplyPatternStep struct {
	PatternSHA string
	TargetPath string
	Params     map[string]string
}

// Kind returns the step kind.
func (ApplyPatternStep) Kind() string { return "apply-pattern" }

// WriteFileStep writes raw content to a file.
type WriteFileStep struct {
	Path    string
	Content []byte
}

// Kind returns the step kind.
func (WriteFileStep) Kind() string { return "write-file" }

// EditPatchStep applies a unified diff patch to a file.
type EditPatchStep struct {
	Path  string
	Patch string
}

// Kind returns the step kind.
func (EditPatchStep) Kind() string { return "edit-patch" }

// emptyTreeSHA is the canonical SHA-1 of git's empty tree object.
const emptyTreeSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// EmptyTreeSHA is the canonical SHA-1 of git's empty tree object.
const EmptyTreeSHA = emptyTreeSHA

// Assembler applies ordered AssemblySteps to a git tree and returns a new tree SHA.
type Assembler struct {
	store *Store
}

// NewAssembler creates an Assembler backed by store.
func NewAssembler(store *Store) *Assembler {
	return &Assembler{store: store}
}

// Assemble materializes inputTreeSHA into a temp directory, runs each step,
// then writes the resulting contents back as a new git tree and returns its SHA.
func (a *Assembler) Assemble(ctx context.Context, inputTreeSHA string, steps []AssemblyStep) (string, error) {
	if a.store == nil {
		return "", fmt.Errorf("assembler store is nil")
	}

	workDir, err := os.MkdirTemp("", "mimic-assembly-*")
	if err != nil {
		return "", fmt.Errorf("create work dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	gitDir := filepath.Join(a.store.repoPath, ".git")

	// Materialize the input tree into the work directory.
	if err := a.materializeTree(ctx, gitDir, workDir, inputTreeSHA); err != nil {
		return "", fmt.Errorf("materialize tree %s: %w", inputTreeSHA, err)
	}

	// Apply each step in order.
	for i, step := range steps {
		if err := a.applyStep(ctx, gitDir, workDir, step); err != nil {
			return "", fmt.Errorf("step %d (%s): %w", i, step.Kind(), err)
		}
	}

	// Build a new tree object from the work directory contents.
	out, err := a.runGit(ctx, workDir, gitDir, "add", "-A")
	if err != nil {
		return "", fmt.Errorf("git add: %w", err)
	}
	_ = out

	out, err = a.runGit(ctx, workDir, gitDir, "write-tree")
	if err != nil {
		return "", fmt.Errorf("git write-tree: %w", err)
	}
	newTreeSHA := strings.TrimSpace(string(out))
	if newTreeSHA == "" {
		return "", fmt.Errorf("git write-tree returned empty SHA")
	}

	return newTreeSHA, nil
}

func (a *Assembler) materializeTree(ctx context.Context, gitDir, workDir, treeSHA string) error {
	if treeSHA == "" {
		return fmt.Errorf("input tree SHA is empty")
	}

	// The empty tree has no content; nothing to checkout.
	if treeSHA == emptyTreeSHA {
		return nil
	}

	// Materialize a tree object (not a commit) via git archive + tar extraction.
	archiveCmd := exec.CommandContext(ctx, "git", "--git-dir", gitDir, "archive", treeSHA)
	tarCmd := exec.CommandContext(ctx, "tar", "-x", "-C", workDir)
	pipe, err := archiveCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("git archive pipe: %w", err)
	}
	tarCmd.Stdin = pipe

	if err := archiveCmd.Start(); err != nil {
		return fmt.Errorf("git archive start: %w", err)
	}
	if out, err := tarCmd.CombinedOutput(); err != nil {
		archiveCmd.Wait()
		return fmt.Errorf("tar extract: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	if err := archiveCmd.Wait(); err != nil {
		return fmt.Errorf("git archive: %w", err)
	}
	return nil
}

func (a *Assembler) applyStep(ctx context.Context, gitDir, workDir string, step AssemblyStep) error {
	switch s := step.(type) {
	case ApplyPatternStep:
		return a.applyPatternStep(gitDir, workDir, s)
	case WriteFileStep:
		return a.applyWriteFileStep(workDir, s)
	case EditPatchStep:
		return a.applyEditPatchStep(ctx, gitDir, workDir, s)
	default:
		return fmt.Errorf("unknown step type %T", step)
	}
}

func (a *Assembler) applyPatternStep(gitDir, workDir string, step ApplyPatternStep) error {
	if step.PatternSHA == "" {
		return fmt.Errorf("pattern SHA is empty")
	}
	if step.TargetPath == "" {
		return fmt.Errorf("target path is empty")
	}

	slot, err := a.store.LoadPattern(step.PatternSHA)
	if err != nil {
		return fmt.Errorf("load pattern %s: %w", step.PatternSHA, err)
	}

	content := renderPatternContent(slot, step.Params)
	fullPath := filepath.Join(workDir, step.TargetPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("mkdir for %s: %w", step.TargetPath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", step.TargetPath, err)
	}
	return nil
}

func renderPatternContent(slot *mesh.TextSlot, params map[string]string) string {
	var content string
	if len(slot.Actions) > 0 {
		content = strings.Join(slot.Actions, "\n")
	} else {
		content = slot.Context
	}
	return interpolateParams(content, params)
}

func interpolateParams(content string, params map[string]string) string {
	for k, v := range params {
		content = strings.ReplaceAll(content, "$"+k, v)
		content = strings.ReplaceAll(content, "${"+k+"}", v)
	}
	return content
}

func (a *Assembler) applyWriteFileStep(workDir string, step WriteFileStep) error {
	if step.Path == "" {
		return fmt.Errorf("path is empty")
	}
	fullPath := filepath.Join(workDir, step.Path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("mkdir for %s: %w", step.Path, err)
	}
	if err := os.WriteFile(fullPath, step.Content, 0644); err != nil {
		return fmt.Errorf("write %s: %w", step.Path, err)
	}
	return nil
}

func (a *Assembler) applyEditPatchStep(ctx context.Context, gitDir, workDir string, step EditPatchStep) error {
	if step.Path == "" {
		return fmt.Errorf("path is empty")
	}
	if step.Patch == "" {
		return nil
	}
	fullPath := filepath.Join(workDir, step.Path)

	// Prefer system patch when available; it handles unified diffs robustly.
	if _, err := exec.LookPath("patch"); err == nil {
		cmd := exec.CommandContext(ctx, "patch", "-p0")
		cmd.Dir = workDir
		cmd.Stdin = strings.NewReader(step.Patch)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("patch %s: %w\n%s", step.Path, err, strings.TrimSpace(string(out)))
		}
		return nil
	}

	// Minimal fallback: apply a unified diff line-by-line.
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", step.Path, err)
	}
	lines := strings.Split(string(data), "\n")
	patched, err := applyUnifiedDiffFallback(lines, step.Patch)
	if err != nil {
		return fmt.Errorf("apply patch to %s: %w", step.Path, err)
	}
	if err := os.WriteFile(fullPath, []byte(strings.Join(patched, "\n")), 0644); err != nil {
		return fmt.Errorf("write %s: %w", step.Path, err)
	}
	return nil
}

func applyUnifiedDiffFallback(lines []string, patch string) ([]string, error) {
	var out []string
	idx := 0
	inHunk := false

	for _, rawLine := range strings.Split(patch, "\n") {
		if strings.HasPrefix(rawLine, "---") || strings.HasPrefix(rawLine, "+++") {
			continue
		}
		if strings.HasPrefix(rawLine, "@@") {
			inHunk = true
			continue
		}
		if !inHunk {
			continue
		}

		switch {
		case strings.HasPrefix(rawLine, "-"):
			// Remove line from current position.
			if idx >= len(lines) {
				return nil, fmt.Errorf("unexpected removal at line %d", idx)
			}
			want := strings.TrimPrefix(rawLine, "-")
			if lines[idx] != want {
				return nil, fmt.Errorf("removal mismatch at line %d: got %q, want %q", idx, lines[idx], want)
			}
			idx++
		case strings.HasPrefix(rawLine, "+"):
			// Insert line before current position.
			out = append(out, strings.TrimPrefix(rawLine, "+"))
		case rawLine == " ":
			// Context line; keep and advance.
			if idx >= len(lines) {
				return nil, fmt.Errorf("unexpected context at line %d", idx)
			}
			out = append(out, lines[idx])
			idx++
		case rawLine == "":
			// Empty patch lines are ignored.
		default:
			// Context line without leading space (tolerate).
			if idx >= len(lines) {
				return nil, fmt.Errorf("unexpected context at line %d", idx)
			}
			out = append(out, lines[idx])
			idx++
		}
	}

	out = append(out, lines[idx:]...)
	return out, nil
}

func (a *Assembler) runGit(ctx context.Context, workDir, gitDir string, args ...string) ([]byte, error) {
	fullArgs := append([]string{"--work-tree", workDir, "--git-dir", gitDir}, args...)
	cmd := exec.CommandContext(ctx, "git", fullArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return out, nil
}
