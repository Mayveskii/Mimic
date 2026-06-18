package cas

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// TreeEntry describes a single entry inside a git tree object.
type TreeEntry struct {
	Mode string
	Type string
	SHA  string
	Path string
}

// Store is a content-addressable object store backed by a git repository.
type Store struct {
	repoPath      string
	patternsIndex []PatternRef
}

// NewStore opens a CAS store backed by the git repository at repoPath.
// It returns an error if repoPath is empty or not a git repository.
func NewStore(repoPath string) (*Store, error) {
	if repoPath == "" {
		return nil, fmt.Errorf("repo path is empty")
	}

	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "--git-dir").Output()
	if err != nil {
		return nil, fmt.Errorf("not a git repository at %q: %w", repoPath, cmdErr(err))
	}
	if strings.TrimSpace(string(out)) == "" {
		return nil, fmt.Errorf("not a git repository at %q", repoPath)
	}

	return &Store{repoPath: repoPath}, nil
}

// StoreBlob writes a blob object and returns its SHA-1 hash.
func (s *Store) StoreBlob(data []byte) (string, error) {
	cmd := exec.Command("git", "-C", s.repoPath, "hash-object", "-w", "--stdin")
	cmd.Stdin = bytes.NewReader(data)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("store blob: %w", cmdErr(err))
	}
	return strings.TrimSpace(string(out)), nil
}

// StoreTree writes a tree object from the supplied entries and returns its hash.
// Entries are sorted by path before being passed to git mktree.
func (s *Store) StoreTree(entries []TreeEntry) (string, error) {
	sorted := make([]TreeEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	var buf bytes.Buffer
	for _, e := range sorted {
		fmt.Fprintf(&buf, "%s %s %s\t%s\n", e.Mode, e.Type, e.SHA, e.Path)
	}

	cmd := exec.Command("git", "-C", s.repoPath, "mktree")
	cmd.Stdin = &buf
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("store tree: %w", cmdErr(err))
	}
	return strings.TrimSpace(string(out)), nil
}

// StoreCommit writes a commit object pointing to treeSHA.
// If parentSHA is empty, a root commit is created.
func (s *Store) StoreCommit(treeSHA, parentSHA, message string) (string, error) {
	if treeSHA == "" {
		return "", fmt.Errorf("tree SHA is empty")
	}
	if message == "" {
		return "", fmt.Errorf("commit message is empty")
	}

	args := []string{"-C", s.repoPath, "commit-tree", treeSHA}
	if parentSHA != "" {
		args = append(args, "-p", parentSHA)
	}
	args = append(args, "-m", message)

	cmd := exec.Command("git", args...)
	cmd.Env = commitEnv()
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("store commit: %w", cmdErr(err))
	}
	return strings.TrimSpace(string(out)), nil
}

// ReadBlob returns the contents of a blob object.
func (s *Store) ReadBlob(sha string) ([]byte, error) {
	if sha == "" {
		return nil, fmt.Errorf("blob SHA is empty")
	}

	cmd := exec.Command("git", "-C", s.repoPath, "cat-file", "-p", sha)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("read blob %s: %w", sha, cmdErr(err))
	}
	return out, nil
}

// ReadTree parses a tree object into its entries.
func (s *Store) ReadTree(sha string) ([]TreeEntry, error) {
	if sha == "" {
		return nil, fmt.Errorf("tree SHA is empty")
	}

	cmd := exec.Command("git", "-C", s.repoPath, "cat-file", "-p", sha)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("read tree %s: %w", sha, cmdErr(err))
	}

	var entries []TreeEntry
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		// Format: "<mode> <type> <sha>\t<path>"
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tree line: %q", line)
		}
		meta := strings.Fields(parts[0])
		if len(meta) != 3 {
			return nil, fmt.Errorf("invalid tree entry metadata: %q", parts[0])
		}
		entries = append(entries, TreeEntry{
			Mode: meta[0],
			Type: meta[1],
			SHA:  meta[2],
			Path: parts[1],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan tree %s: %w", sha, err)
	}

	return entries, nil
}

// cmdErr extracts a useful error from an exec.ExitError.
func cmdErr(err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		if stderr != "" {
			return fmt.Errorf("%s", stderr)
		}
	}
	return err
}

// commitEnv returns an environment that supplies default author/committer
// identities when the calling process does not already have them set.
func commitEnv() []string {
	env := os.Environ()
	if os.Getenv("GIT_AUTHOR_NAME") == "" {
		env = append(env, "GIT_AUTHOR_NAME=Mimic CAS")
	}
	if os.Getenv("GIT_AUTHOR_EMAIL") == "" {
		env = append(env, "GIT_AUTHOR_EMAIL=mimic@localhost")
	}
	if os.Getenv("GIT_COMMITTER_NAME") == "" {
		env = append(env, "GIT_COMMITTER_NAME=Mimic CAS")
	}
	if os.Getenv("GIT_COMMITTER_EMAIL") == "" {
		env = append(env, "GIT_COMMITTER_EMAIL=mimic@localhost")
	}
	return env
}
