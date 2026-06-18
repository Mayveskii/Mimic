package sandbox

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Mayveskii/Mimic/internal/cas"
)

// SessionRef tracks per-session state as a git ref backed by a CAS store.
// The ref lives under refs/sessions/{uuid} and may point to either a tree
// (initial baseline) or a commit (after advances).
type SessionRef struct {
	store *cas.Store
}

// NewSessionRef creates a SessionRef backed by the supplied CAS store.
func NewSessionRef(store *cas.Store) *SessionRef {
	return &SessionRef{store: store}
}

// Create creates refs/sessions/{uuid} pointing to baseTreeSHA.
func (s *SessionRef) Create(uuid, baseTreeSHA string) error {
	if uuid == "" {
		return fmt.Errorf("uuid is empty")
	}
	if baseTreeSHA == "" {
		return fmt.Errorf("base tree SHA is empty")
	}

	cmd := exec.Command("git", "-C", s.store.RepoPath(), "update-ref", s.refName(uuid), baseTreeSHA)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("update-ref: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Advance creates a commit from treeSHA, chained from the current ref target if
// it is a commit, and updates refs/sessions/{uuid} to the new commit. If the ref
// is missing or points to a non-commit (e.g. a baseline tree), a root commit is
// created. It returns the new commit SHA.
func (s *SessionRef) Advance(uuid, treeSHA, message string) (string, error) {
	if uuid == "" {
		return "", fmt.Errorf("uuid is empty")
	}
	if treeSHA == "" {
		return "", fmt.Errorf("tree SHA is empty")
	}
	if message == "" {
		return "", fmt.Errorf("message is empty")
	}

	parent, _ := s.currentCommit(uuid)
	commitSHA, err := s.store.StoreCommit(treeSHA, parent, message)
	if err != nil {
		return "", fmt.Errorf("store commit: %w", err)
	}

	cmd := exec.Command("git", "-C", s.store.RepoPath(), "update-ref", s.refName(uuid), commitSHA)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("update-ref: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return commitSHA, nil
}

// Rewind updates refs/sessions/{uuid} to point to treeSHA.
func (s *SessionRef) Rewind(uuid, treeSHA string) error {
	if uuid == "" {
		return fmt.Errorf("uuid is empty")
	}
	if treeSHA == "" {
		return fmt.Errorf("tree SHA is empty")
	}

	cmd := exec.Command("git", "-C", s.store.RepoPath(), "update-ref", s.refName(uuid), treeSHA)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("update-ref: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// History returns commit SHAs for refs/sessions/{uuid} in reverse chronological
// order (newest first), using git log.
func (s *SessionRef) History(uuid string) ([]string, error) {
	if uuid == "" {
		return nil, fmt.Errorf("uuid is empty")
	}

	cmd := exec.Command("git", "-C", s.store.RepoPath(), "log", s.refName(uuid), "--format=%H")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var shas []string
	for _, line := range lines {
		sha := strings.TrimSpace(line)
		if sha != "" {
			shas = append(shas, sha)
		}
	}
	return shas, nil
}

func (s *SessionRef) refName(uuid string) string {
	return fmt.Sprintf("refs/sessions/%s", uuid)
}

// currentCommit resolves the ref to a commit SHA, returning an error if the ref
// is missing or does not point to a commit object.
func (s *SessionRef) currentCommit(uuid string) (string, error) {
	out, err := exec.Command("git", "-C", s.store.RepoPath(), "rev-parse", "--verify", s.refName(uuid)+"^{commit}").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
