package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
