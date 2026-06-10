package config

// RepoConfig is the repo-scoped configuration loaded from .mimic/<repo>/config.yaml
type RepoConfig struct {
	Repo      string         `yaml:"repo"`
	Models    ModelProfiles  `yaml:"models"`
	Budget    BudgetConfig   `yaml:"budget"`
	Sandbox   SandboxConfig  `yaml:"sandbox"`
	Personas  []string       `yaml:"personas,omitempty"`  // paths to persona files
	Primers   []string       `yaml:"primers,omitempty"`   // paths to primer files
	Waivers   []string       `yaml:"waivers,omitempty"`   // paths to waiver files
}

// ModelProfiles maps model categories to concrete model IDs.
type ModelProfiles struct {
	Local   string `yaml:"local"`   // fastest_good equivalent
	Medium  string `yaml:"medium"`  // balanced
	Top     string `yaml:"top"`     // best_code / frontier_best
}

// BudgetConfig defines session-level budget constraints.
type BudgetConfig struct {
	MaxTokens      int `yaml:"max_tokens"`
	MaxIterations  int `yaml:"max_iterations"`
	MaxTimeSeconds int `yaml:"max_time_seconds"`
}

// SandboxConfig defines sandbox behavior.
type SandboxConfig struct {
	WorktreePrefix string `yaml:"worktree_prefix"`
	AutoRollback   bool   `yaml:"auto_rollback"`
	CollectProof   bool   `yaml:"collect_proof"`
}

// DefaultRepoConfig returns a sensible default config.
func DefaultRepoConfig() *RepoConfig {
	return &RepoConfig{
		Models: ModelProfiles{
			Local:  "qwen/qwen3-235b",
			Medium: "moonshotai/kimi-k2.6",
			Top:    "minimaxai/minimax-m2.7",
		},
		Budget: BudgetConfig{
			MaxTokens:      100000,
			MaxIterations:  10,
			MaxTimeSeconds: 600,
		},
		Sandbox: SandboxConfig{
			WorktreePrefix: "",
			AutoRollback:   true,
			CollectProof:   true,
		},
	}
}
