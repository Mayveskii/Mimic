package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mayveskii/Mimic/internal/cgo"
	"github.com/Mayveskii/Mimic/internal/mcp"
	"github.com/Mayveskii/Mimic/internal/tool/exa"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func checkUpdate() {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/Mayveskii/Mimic/releases/latest")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(version, "v")
	if latest != "" && latest != current && current != "dev" {
		fmt.Fprintf(os.Stderr, "mimic: update available: v%s → v%s\n", current, latest)
		fmt.Fprintf(os.Stderr, "mimic: run: curl -sSL https://raw.githubusercontent.com/Mayveskii/Mimic/main/install.sh | bash\n")
	}
}

func selfUpdate() error {
	fmt.Println("mimic: checking for updates...")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/Mayveskii/Mimic/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name        string `json:"name"`
			DownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode release: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(version, "v")
	if latest == "" || latest == current {
		fmt.Println("mimic: already up to date")
		return nil
	}

	// Find asset
	targetAsset := fmt.Sprintf("mimic_v%s_linux_amd64.tar.gz", latest)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == targetAsset {
			downloadURL = asset.DownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("asset %s not found in release %s", targetAsset, release.TagName)
	}

	fmt.Printf("mimic: updating v%s → v%s\n", current, latest)

	// Determine executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to locate executable: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Download to temp file
	tmpDir, err := os.MkdirTemp("", "mimic-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, targetAsset)
	out, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	resp2, err := client.Get(downloadURL)
	if err != nil {
		out.Close()
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		out.Close()
		return fmt.Errorf("download failed: HTTP %d", resp2.StatusCode)
	}

	_, err = io.Copy(out, resp2.Body)
	out.Close()
	if err != nil {
		return fmt.Errorf("failed to save download: %w", err)
	}

	// Extract
	if err := exec.Command("tar", "-xzf", tarPath, "-C", tmpDir).Run(); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	newBinary := filepath.Join(tmpDir, "mimic")
	if _, err := os.Stat(newBinary); err != nil {
		return fmt.Errorf("extracted binary not found: %w", err)
	}

	// Make executable
	if err := os.Chmod(newBinary, 0755); err != nil {
		return fmt.Errorf("failed to chmod new binary: %w", err)
	}

	// Replace current binary (os.Rename works on Linux even for running binary)
	if err := os.Rename(newBinary, exePath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("mimic: updated to v%s successfully\n", latest)
	return nil
}

func determineWorkingDir() string {
	// Priority 1: explicit env override (for containers, systemd, etc.)
	if dir := os.Getenv("MIMIC_WORKING_DIR"); dir != "" {
		if abs, err := filepath.Abs(dir); err == nil {
			return abs
		}
		return dir
	}

	// Priority 2: directory of the executable
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		if filepath.Base(exeDir) == "bin" {
			if parent := filepath.Dir(exeDir); parent != "" {
				return parent
			}
		}
		return exeDir
	}

	// Priority 3: current working directory at startup
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}

	return "."
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--check-update" {
		checkUpdate()
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "update" {
		if err := selfUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "mimic: update failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("mimic %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "serve" {
		workingDir := determineWorkingDir()

		// Check for updates (non-blocking, best-effort)
		go checkUpdate()

		// Load .env files BEFORE any config reading (so EXA_API_KEY etc. are available)
		exa.EnsureEnvLoaded()

		// Initialize c-core
		if err := cgo.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "mimic: failed to initialize c-core: %v\n", err)
			os.Exit(1)
		}
		defer cgo.Shutdown()

		// Configure mesh data paths from env or working dir
		meshDir := os.Getenv("MIMIC_MESH_DIR")
		if meshDir == "" {
			meshDir = filepath.Join(workingDir, "data", "mesh", "graphs")
		}
		embedEndpoint := os.Getenv("MIMIC_EMBED_ENDPOINT")
		if embedEndpoint == "" {
			embedEndpoint = "http://localhost:1137"
		}

		// Detect TCP mode from env or flag
		tcpAddr := os.Getenv("MIMIC_TCP_ADDR")
		for i, arg := range os.Args {
			if arg == "--tcp" && i+1 < len(os.Args) {
				tcpAddr = os.Args[i+1]
				break
			}
		}

		fmt.Fprintf(os.Stderr, "mimic: workingDir=%s meshDir=%s embed=%s\n", workingDir, meshDir, embedEndpoint)

		if tcpAddr != "" {
			// Pre-initialize server once so mesh is loaded in memory;
			// clone per TCP connection via WithTransport.
			stdioTransport := mcp.NewStdioTransport()
			templateServer := mcp.NewServer(stdioTransport, workingDir, meshDir, embedEndpoint)
			if err := mcp.ServeTCP(tcpAddr, templateServer); err != nil {
				fmt.Fprintf(os.Stderr, "mimic: tcp server error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Start MCP server over stdio with deterministic working directory
		transport := mcp.NewStdioTransport()
		server := mcp.NewServer(transport, workingDir, meshDir, embedEndpoint)
		if err := server.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "mimic: server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("mimic — MCP server for AI agent tool orchestration")
	fmt.Println("Usage: mimic serve")
	fmt.Println("       mimic serve --tcp :1337")
	fmt.Println("See specs/ for project documentation and architecture.")
}
