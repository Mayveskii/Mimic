package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRepoConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "rtk")
	_ = os.MkdirAll(repoPath, 0755)

	cfg, err := LoadRepoConfig(repoPath)
	if err != nil {
		t.Fatalf("LoadRepoConfig: %v", err)
	}
	if cfg.Repo != "rtk" {
		t.Fatalf("expected repo=rtk, got %q", cfg.Repo)
	}
	if cfg.Budget.MaxTokens != 100000 {
		t.Fatalf("expected max_tokens=100000, got %d", cfg.Budget.MaxTokens)
	}
	if cfg.Models.Local != "qwen/qwen3-235b" {
		t.Fatalf("unexpected local model: %q", cfg.Models.Local)
	}
}

func TestLoadRepoConfig_Custom(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "rtk")
	configDir := filepath.Join(repoPath, ".mimic", "rtk")
	_ = os.MkdirAll(configDir, 0755)

	yaml := `repo: rtk
models:
  local: custom/local
  medium: custom/medium
  top: custom/top
budget:
  max_tokens: 50000
  max_iterations: 5
  max_time_seconds: 300
sandbox:
  auto_rollback: false
  collect_proof: false
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadRepoConfig(repoPath)
	if err != nil {
		t.Fatalf("LoadRepoConfig: %v", err)
	}
	if cfg.Budget.MaxTokens != 50000 {
		t.Fatalf("expected max_tokens=50000, got %d", cfg.Budget.MaxTokens)
	}
	if cfg.Budget.MaxIterations != 5 {
		t.Fatalf("expected max_iterations=5, got %d", cfg.Budget.MaxIterations)
	}
	if cfg.Sandbox.AutoRollback == nil || *cfg.Sandbox.AutoRollback {
		t.Fatal("expected auto_rollback=false")
	}
	if cfg.Models.Local != "custom/local" {
		t.Fatalf("expected local model custom/local, got %q", cfg.Models.Local)
	}
}

func TestSaveRepoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "rtk")

	cfg := &RepoConfig{
		Repo:   "rtk",
		Models: ModelProfiles{Local: "qwen", Medium: "kimi", Top: "minimax"},
		Budget: BudgetConfig{MaxTokens: 100000, MaxIterations: 10, MaxTimeSeconds: 600},
	}

	if err := SaveRepoConfig(repoPath, cfg); err != nil {
		t.Fatalf("SaveRepoConfig: %v", err)
	}

	configPath := filepath.Join(repoPath, ".mimic", "rtk", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Round-trip load
	cfg2, err := LoadRepoConfig(repoPath)
	if err != nil {
		t.Fatalf("LoadRepoConfig after save: %v", err)
	}
	if cfg2.Models.Local != "qwen" {
		t.Fatalf("round-trip failed: expected qwen, got %q", cfg2.Models.Local)
	}
}
