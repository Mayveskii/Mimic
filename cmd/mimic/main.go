package main

import (
	"fmt"
	"os"

	"github.com/Mayveskii/Mimic/internal/cgo"
	"github.com/Mayveskii/Mimic/internal/mcp"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		// Initialize c-core
		if err := cgo.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "mimic: failed to initialize c-core: %v\n", err)
			os.Exit(1)
		}
		defer cgo.Shutdown()

		// Start MCP server over stdio
		transport := mcp.NewStdioTransport()
		server := mcp.NewServer(transport)
		if err := server.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "mimic: server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("mimic — MCP server for AI agent tool orchestration")
	fmt.Println("Usage: mimic serve")
	fmt.Println("See specs/ for project documentation and architecture.")
}
