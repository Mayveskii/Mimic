package exa

import (
	"os"
	"strconv"
	"time"
)

// LoadConfigFromEnv reads EXA_* variables from the environment.
// Invariant: no default for APIKey (empty = disabled).
// Source: Mayveskii/exa-mcp-server (rate_limit_management).
func LoadConfigFromEnv() Config {
	cfg := Config{
		APIKey:             os.Getenv("EXA_API_KEY"),
		BaseURL:            getEnvDefault("EXA_BASE_URL", "https://api.exa.ai"),
		MaxResults:         getEnvIntDefault("EXA_MAX_RESULTS", 10),
		TimeoutMs:          getEnvIntDefault("EXA_TIMEOUT_MS", 30000),
		RetryMax:           getEnvIntDefault("EXA_RETRY_MAX", 3),
		RetryBackoffBaseMs: getEnvIntDefault("EXA_RETRY_BACKOFF_BASE_MS", 1000),
	}
	if cfg.MaxResults < 1 || cfg.MaxResults > 100 {
		cfg.MaxResults = 10
	}
	if cfg.TimeoutMs < 1000 {
		cfg.TimeoutMs = 30000
	}
	if cfg.RetryMax < 0 || cfg.RetryMax > 10 {
		cfg.RetryMax = 3
	}
	if cfg.RetryBackoffBaseMs < 100 {
		cfg.RetryBackoffBaseMs = 1000
	}
	return cfg
}

// Disabled returns true if the Exa client should not be used.
func (c Config) Disabled() bool {
	return c.APIKey == "" || c.BaseURL == ""
}

// Timeout returns the per-request timeout as time.Duration.
func (c Config) Timeout() time.Duration {
	return time.Duration(c.TimeoutMs) * time.Millisecond
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvIntDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
