package config

// RepoConfig is the repo-scoped configuration loaded from .mimic/<repo>/config.yaml
type RepoConfig struct {
	Repo      string                    `yaml:"repo"`
	Models    ModelProfiles             `yaml:"models"`
	Providers map[string]ProviderConfig `yaml:"providers,omitempty"`
	Budget    BudgetConfig              `yaml:"budget"`
	Sandbox   SandboxConfig             `yaml:"sandbox"`
	Value     ValueConfig               `yaml:"value,omitempty"`
	Personas  []string                  `yaml:"personas,omitempty"` // paths to persona files
	Primers   []string                  `yaml:"primers,omitempty"`  // paths to primer files
	Waivers   []string                  `yaml:"waivers,omitempty"`  // paths to waiver files
}

// ModelProfiles maps model categories to concrete model IDs.
type ModelProfiles struct {
	Local   string        `yaml:"local"`  // fastest_good equivalent
	Medium  string        `yaml:"medium"` // balanced
	Top     string        `yaml:"top"`    // best_code / frontier_best
	Cascade CascadeConfig `yaml:"cascade,omitempty"`
}

// CascadeConfig controls the local -> medium -> top escalation behavior.
type CascadeConfig struct {
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	MaxEscalations      int     `yaml:"max_escalations"`
}

// ProviderConfig describes an LLM API provider.
type ProviderConfig struct {
	Endpoint  string `yaml:"endpoint"`
	EnvKey    string `yaml:"env_key"`
	TimeoutMs int    `yaml:"timeout_ms"`
	RetryMax  int    `yaml:"retry_max"`
}

// BudgetConfig defines session-level budget constraints.
type BudgetConfig struct {
	MaxTokens      int `yaml:"max_tokens"`
	MaxIterations  int `yaml:"max_iterations"`
	MaxTimeSeconds int `yaml:"max_time_seconds"`
}

// SandboxConfig defines sandbox behavior.
// Boolean fields are pointers so we can distinguish "not set" from "explicitly false".
type SandboxConfig struct {
	WorktreePrefix string `yaml:"worktree_prefix"`
	AutoRollback   *bool  `yaml:"auto_rollback,omitempty"`
	CollectProof   *bool  `yaml:"collect_proof,omitempty"`
}

// ValueConfig defines how finding value is estimated for cost/value reports.
type ValueConfig struct {
	ValuePerSeverity map[int]float64 `yaml:"value_per_severity,omitempty"`
}

func boolPtr(b bool) *bool { return &b }

// DefaultRepoConfig returns a sensible default config.
func DefaultRepoConfig() *RepoConfig {
	return &RepoConfig{
		Models: ModelProfiles{
			Local:  "qwen/qwen3-235b",
			Medium: "moonshotai/kimi-k2.6",
			Top:    "minimaxai/minimax-m2.7",
			Cascade: CascadeConfig{
				ConfidenceThreshold: 0.85,
				MaxEscalations:      2,
			},
		},
		Providers: map[string]ProviderConfig{
			"gonkagate": {
				Endpoint:  "https://api.gonkagate.com/v1",
				EnvKey:    "GONKAGATE_API_KEY",
				TimeoutMs: 120000,
				RetryMax:  3,
			},
		},
		Budget: BudgetConfig{
			MaxTokens:      100000,
			MaxIterations:  10,
			MaxTimeSeconds: 600,
		},
		Sandbox: SandboxConfig{
			WorktreePrefix: "",
			AutoRollback:   boolPtr(true),
			CollectProof:   boolPtr(true),
		},
		Value: ValueConfig{
			ValuePerSeverity: map[int]float64{
				1: 10.0,
				2: 50.0,
				3: 200.0,
				4: 1000.0,
				5: 5000.0,
			},
		},
	}
}
