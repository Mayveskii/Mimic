package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Mayveskii/Mimic/internal/cgo"
	"github.com/Mayveskii/Mimic/internal/mcp"
	"github.com/Mayveskii/Mimic/internal/tool/exa"
)

func determineWorkingDir() string {
	// Priority 1: explicit env override (for containers, systemd, etc.)
	if dir := os.Getenv("MIMIC_WORKING_DIR"); dir != "" {
		if abs, err := filepath.Abs(dir); err == nil {
			return abs
		}
		return dir
	}

	// Priority 2: directory of the executable
	// If mimic is installed at /opt/mimic/bin/mimic, workingDir = /opt/mimic
	// If installed at /usr/local/bin/mimic, workingDir = /usr/local/bin
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		// If binary lives in a bin/ directory, use the parent as project root
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

	// Fallback: current directory
	return "."
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		workingDir := determineWorkingDir()

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
