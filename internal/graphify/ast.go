package graphify

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Node represents a symbol in the repo graph.
type Node struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Type     string   `json:"type"`  // func, type, var, import
	Edges    []string `json:"edges"` // IDs of connected nodes
	PageRank float64  `json:"page_rank"`
}

// Graph is the repo knowledge graph.
// Behavior source: Aider repo map + graphify IDF-weighted graph.
type Graph struct {
	Nodes map[string]*Node
}

// NewGraph creates an empty graph.
func NewGraph() *Graph {
	return &Graph{Nodes: make(map[string]*Node)}
}

// BuildFromRepo scans a repository and extracts symbols into a graph.
func (g *Graph) BuildFromRepo(repoPath string) error {
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(path, "/.git/") || strings.Contains(path, "/vendor/") {
			return nil
		}

		switch filepath.Ext(path) {
		case ".go":
			return g.parseGoFile(path)
		case ".rs":
			return g.parseRustFile(path)
		case ".py":
			return g.parsePythonFile(path)
		}
		return nil
	})
}

// parseGoFile extracts functions, types, and imports from a Go file.
func (g *Graph) parseGoFile(path string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil // skip unparseable files
	}

	// Imports
	for _, imp := range f.Imports {
		name := strings.Trim(imp.Path.Value, `"`)
		nodeID := fmt.Sprintf("import:%s", name)
		if _, ok := g.Nodes[nodeID]; !ok {
			g.Nodes[nodeID] = &Node{
				ID:   nodeID,
				Name: name,
				File: path,
				Line: fset.Position(imp.Pos()).Line,
				Type: "import",
			}
		}
	}

	// Functions and types
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			nodeID := fmt.Sprintf("func:%s:%s", path, x.Name.Name)
			node := &Node{
				ID:   nodeID,
				Name: x.Name.Name,
				File: path,
				Line: fset.Position(x.Pos()).Line,
				Type: "func",
			}
			g.Nodes[nodeID] = node

			// Simple edge detection: function calls other functions in same file
			ast.Inspect(x.Body, func(bodyNode ast.Node) bool {
				if call, ok := bodyNode.(*ast.CallExpr); ok {
					if ident, ok := call.Fun.(*ast.Ident); ok {
						targetID := fmt.Sprintf("func:%s:%s", path, ident.Name)
						if targetID != nodeID {
							node.Edges = append(node.Edges, targetID)
						}
					}
				}
				return true
			})

		case *ast.TypeSpec:
			nodeID := fmt.Sprintf("type:%s:%s", path, x.Name.Name)
			g.Nodes[nodeID] = &Node{
				ID:   nodeID,
				Name: x.Name.Name,
				File: path,
				Line: fset.Position(x.Pos()).Line,
				Type: "type",
			}
		}
		return true
	})

	return nil
}

// parseRustFile uses regex fallback for Rust symbols.
func (g *Graph) parseRustFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "fn ") {
			name := strings.Fields(line)[1]
			name = strings.TrimRight(name, "(<{")
			nodeID := fmt.Sprintf("func:%s:%s", path, name)
			g.Nodes[nodeID] = &Node{
				ID:   nodeID,
				Name: name,
				File: path,
				Line: i + 1,
				Type: "func",
			}
		}
		if strings.HasPrefix(line, "struct ") || strings.HasPrefix(line, "enum ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				name := fields[1]
				nodeID := fmt.Sprintf("type:%s:%s", path, name)
				g.Nodes[nodeID] = &Node{
					ID:   nodeID,
					Name: name,
					File: path,
					Line: i + 1,
					Type: "type",
				}
			}
		}
	}
	return nil
}

// parsePythonFile uses regex fallback for Python symbols.
func (g *Graph) parsePythonFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "def ") {
			name := strings.Fields(line)[1]
			name = strings.TrimSuffix(name, "(")
			name = strings.TrimSuffix(name, ":")
			nodeID := fmt.Sprintf("func:%s:%s", path, name)
			g.Nodes[nodeID] = &Node{
				ID:   nodeID,
				Name: name,
				File: path,
				Line: i + 1,
				Type: "func",
			}
		}
		if strings.HasPrefix(line, "class ") {
			name := strings.Fields(line)[1]
			name = strings.TrimSuffix(name, "(")
			name = strings.TrimSuffix(name, ":")
			nodeID := fmt.Sprintf("type:%s:%s", path, name)
			g.Nodes[nodeID] = &Node{
				ID:   nodeID,
				Name: name,
				File: path,
				Line: i + 1,
				Type: "type",
			}
		}
	}
	return nil
}

// ComputePageRank runs a simple PageRank algorithm on the graph.
func (g *Graph) ComputePageRank(iterations int, damping float64) {
	if len(g.Nodes) == 0 {
		return
	}
	base := (1.0 - damping) / float64(len(g.Nodes))

	// Initialize
	for _, node := range g.Nodes {
		node.PageRank = 1.0 / float64(len(g.Nodes))
	}

	for i := 0; i < iterations; i++ {
		newRanks := make(map[string]float64)
		for id := range g.Nodes {
			newRanks[id] = base
		}

		for _, node := range g.Nodes {
			if len(node.Edges) == 0 {
				continue
			}
			share := node.PageRank * damping / float64(len(node.Edges))
			for _, targetID := range node.Edges {
				if _, ok := g.Nodes[targetID]; ok {
					newRanks[targetID] += share
				}
			}
		}

		for id, rank := range newRanks {
			g.Nodes[id].PageRank = rank
		}
	}
}

// TopNodes returns the N highest PageRank nodes.
func (g *Graph) TopNodes(n int) []*Node {
	var all []*Node
	for _, node := range g.Nodes {
		all = append(all, node)
	}

	// Simple bubble sort for small N
	for i := 0; i < len(all); i++ {
		for j := i + 1; j < len(all); j++ {
			if all[j].PageRank > all[i].PageRank {
				all[i], all[j] = all[j], all[i]
			}
		}
	}

	if n > len(all) {
		n = len(all)
	}
	return all[:n]
}
