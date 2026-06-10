package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadRepoConfig loads the repo-scoped configuration from .mimic/<repo>/config.yaml.
// If the file does not exist, returns DefaultRepoConfig with repo name filled in.
func LoadRepoConfig(repoPath string) (*RepoConfig, error) {
	repoName := filepath.Base(repoPath)
	configPath := filepath.Join(repoPath, ".mimic", repoName, "config.yaml")

	// If config does not exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := DefaultRepoConfig()
		cfg.Repo = repoName
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", configPath, err)
	}

	cfg := &RepoConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", configPath, err)
	}

	if cfg.Repo == "" {
		cfg.Repo = repoName
	}

	// Apply defaults for missing fields
	defaults := DefaultRepoConfig()
	if cfg.Models.Local == "" {
		cfg.Models.Local = defaults.Models.Local
	}
	if cfg.Models.Medium == "" {
		cfg.Models.Medium = defaults.Models.Medium
	}
	if cfg.Models.Top == "" {
		cfg.Models.Top = defaults.Models.Top
	}
	if cfg.Budget.MaxTokens == 0 {
		cfg.Budget.MaxTokens = defaults.Budget.MaxTokens
	}
	if cfg.Budget.MaxIterations == 0 {
		cfg.Budget.MaxIterations = defaults.Budget.MaxIterations
	}
	if cfg.Budget.MaxTimeSeconds == 0 {
		cfg.Budget.MaxTimeSeconds = defaults.Budget.MaxTimeSeconds
	}

	return cfg, nil
}

// SaveRepoConfig writes the repo config to disk.
func SaveRepoConfig(repoPath string, cfg *RepoConfig) error {
	repoName := filepath.Base(repoPath)
	configDir := filepath.Join(repoPath, ".mimic", repoName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("mkdir config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
