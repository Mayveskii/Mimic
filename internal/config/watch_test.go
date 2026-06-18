package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher_HotReload(t *testing.T) {
	tmpDir := t.TempDir()
	repoName := filepath.Base(tmpDir)
	configDir := filepath.Join(tmpDir, ".mimic", repoName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")

	initial := `repo: testrepo
models:
  local: qwen/qwen3-235b
  medium: moonshotai/kimi-k2.6
  top: minimaxai/minimax-m2.7
budget:
  max_tokens: 1000
`
	if err := os.WriteFile(configPath, []byte(initial), 0644); err != nil {
		t.Fatalf("write initial config: %v", err)
	}

	updated := make(chan *RepoConfig, 1)
	w, err := NewWatcher(tmpDir, func(cfg *RepoConfig) {
		updated <- cfg
	})
	if err != nil {
		t.Fatalf("create watcher: %v", err)
	}
	defer w.Stop()

	if w.Config().Budget.MaxTokens != 1000 {
		t.Fatalf("expected initial max_tokens 1000, got %d", w.Config().Budget.MaxTokens)
	}

	modified := `repo: testrepo
models:
  local: qwen/qwen3-235b
  medium: moonshotai/kimi-k2.6
  top: minimaxai/minimax-m2.7
budget:
  max_tokens: 2000
`
	if err := os.WriteFile(configPath, []byte(modified), 0644); err != nil {
		t.Fatalf("write modified config: %v", err)
	}

	select {
	case cfg := <-updated:
		if cfg.Budget.MaxTokens != 2000 {
			t.Errorf("expected updated max_tokens 2000, got %d", cfg.Budget.MaxTokens)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for config reload")
	}
}
