package graphify

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildFromRepo(t *testing.T) {
	// Create a temporary mock repo
	dir := t.TempDir()
	goFile := filepath.Join(dir, "main.go")
	goContent := `package main

import "fmt"

func main() {
	fmt.Println("hello")
	greet()
}

func greet() {}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	g := NewGraph()
	if err := g.BuildFromRepo(dir); err != nil {
		t.Fatal(err)
	}

	// Should have: 2 funcs + 1 import
	if len(g.Nodes) < 2 {
		t.Fatalf("expected at least 2 nodes, got %d", len(g.Nodes))
	}

	// Check main function exists
	found := false
	for _, node := range g.Nodes {
		if node.Name == "main" && node.Type == "func" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'main' function node")
	}

	g.ComputePageRank(20, 0.85)
	top := g.TopNodes(3)
	if len(top) == 0 {
		t.Error("expected some top nodes after PageRank")
	}
}
