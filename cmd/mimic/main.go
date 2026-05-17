package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		fmt.Fprintln(os.Stderr, "mimic: MCP server not yet implemented — see specs/ for architecture")
		os.Exit(1)
	}
	fmt.Println("mimic — MCP server for AI agent tool orchestration")
	fmt.Println("See specs/ for project documentation and architecture.")
}
