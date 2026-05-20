package exa

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// loadEnvFiles attempts to read .env files and set variables for the current process.
// Searches: 1) ~/.mimic/.env  2) $(pwd)/.env  3) $(binary_dir)/.env
func loadEnvFiles() {
	paths := []string{}

	// User home
	home, _ := os.UserHomeDir()
	if home != "" {
		paths = append(paths, filepath.Join(home, ".mimic", ".env"))
	}

	// Current working directory
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".env"))
	}

	// Directory of mimic binary
	if ex, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(ex), ".env"))
	}

	for _, p := range paths {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			parseEnvFile(p)
		}
	}
}

// parseEnvFile reads KEY=VALUE lines and sets them via os.Setenv ONLY if not already set.
func parseEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		// Remove surrounding quotes
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		// Only set if not already present in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// EnsureEnvLoaded is called once before any config loading to populate environment.
func EnsureEnvLoaded() {
	loadEnvFiles()
}

// Status returns a diagnostic map for the Exa integration.
func Status() map[string]interface{} {
	key := os.Getenv("EXA_API_KEY")
	masked := ""
	if len(key) > 8 {
		masked = key[:4] + "..." + key[len(key)-4:]
	} else if key != "" {
		masked = "***"
	}
	return map[string]interface{}{
		"enabled":     key != "",
		"api_key_set": key != "",
		"api_key":     masked,
		"base_url":    os.Getenv("EXA_BASE_URL"),
	}
}
