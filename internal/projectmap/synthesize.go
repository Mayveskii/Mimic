package projectmap

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

// SynthesizeWorkspaceGraph generates a .graph.gob from the current workspace.
// It reads symbols from the project map, generates embeddings via the embed service,
// and creates a mesh-compatible graph at .mimic/workspace.graph.gob.
type WorkspaceSynthesizer struct {
	pm         *ProjectMap
	embedFn    mesh.EmbedFunc // func(text) -> int8[384]
	outputPath string
}

// NewSynthesizer creates a synthesizer for a workspace.
func NewSynthesizer(pm *ProjectMap, embedFn mesh.EmbedFunc) *WorkspaceSynthesizer {
	return &WorkspaceSynthesizer{
		pm:         pm,
		embedFn:    embedFn,
		outputPath: filepath.Join(pm.root, ".mimic", "workspace.graph.gob"),
	}
}

// Synthesize scans the projectmap and writes a mesh graph.
func (ws *WorkspaceSynthesizer) Synthesize() (*mesh.SemanticMap, error) {
	// Collect all symbols from the index
	syms, err := ws.allSymbols()
	if err != nil {
		return nil, fmt.Errorf("collect symbols: %w", err)
	}

	// Build graph
	graph := &mesh.InvariantGraph{
		Version:   "1.0",
		Domain:    filepath.Base(ws.pm.root),
		UpdatedAt: time.Now(),
	}

	// Create nodes from top-level symbols (packages, exported funcs/types)
	seen := make(map[string]bool)
	for _, s := range syms {
		if seen[s.Name] {
			continue
		}
		if !isSignificantSymbol(s) {
			continue
		}
		seen[s.Name] = true

		description := fmt.Sprintf("%s %s in %s", s.Type, s.Name, s.File)
		if s.Signature != "" {
			description += " | " + s.Signature
		}

		node := mesh.InvariantNode{
			ID:         hashID(description),
			Invariant:  description,
			Domain:     filepath.Base(ws.pm.root),
			SourceRepo: ws.pm.root,
			Task:       description,
			CreatedAt:  time.Now(),
		}

		// Embed the description
		emb := ws.embedFn(description)
		node.EmbedInt8 = emb
		node.EmbedRaw = emb[:]

		graph.Nodes = append(graph.Nodes, node)
	}

	// Also add file-level nodes for key files (README, Makefile, main entrypoints)
	files, err := ws.pm.ListFilesByLang("go", 50)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !isKeyFile(f) {
			continue
		}
		id := hashID(f)
		if seen[id] {
			continue
		}
		seen[id] = true

		description := fmt.Sprintf("Key file: %s", f)
		node := mesh.InvariantNode{
			ID:         id,
			Invariant:  description,
			Domain:     filepath.Base(ws.pm.root),
			SourceRepo: ws.pm.root,
			Task:       description,
			CreatedAt:  time.Now(),
		}
		emb := ws.embedFn(description)
		node.EmbedInt8 = emb
		node.EmbedRaw = emb[:]
		graph.Nodes = append(graph.Nodes, node)
	}

	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("no significant symbols found in workspace")
	}

	// Write .graph.gob
	_ = os.MkdirAll(filepath.Dir(ws.outputPath), 0755)
	f, err := os.Create(ws.outputPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := gob.NewEncoder(f).Encode(graph); err != nil {
		return nil, err
	}

	// Convert to SemanticMap for immediate use
	sm, err := mesh.LoadGraphBinary(ws.outputPath)
	if err != nil {
		return nil, fmt.Errorf("load synthesized graph: %w", err)
	}

	return sm, nil
}

// allSymbols fetches all indexed symbols.
func (ws *WorkspaceSynthesizer) allSymbols() ([]Symbol, error) {
	// Query for common prefixes to get all symbols
	// This is a hack — ideally we'd SELECT * FROM symbols
	var all []Symbol
	for _, prefix := range []string{"", "A", "B", "C", "D", "E", "F", "G", "H", "I",
		"J", "K", "L", "M", "N", "O", "P", "Q", "R", "S",
		"T", "U", "V", "W", "X", "Y", "Z"} {
		ps, err := ws.pm.QuerySymbol(prefix)
		if err != nil {
			return nil, err
		}
		all = append(all, ps...)
	}
	return all, nil
}

func isSignificantSymbol(s Symbol) bool {
	// Skip unexported, short names, internals
	if len(s.Name) < 3 {
		return false
	}
	if s.Type == "var" || s.Type == "const" {
		return false
	}
	return true
}

func isKeyFile(f string) bool {
	base := filepath.Base(f)
	switch base {
	case "main.go", "cmd.go", "server.go", "client.go",
		"README.md", "README", "Makefile", "Dockerfile",
		"go.mod", "go.sum", "main_test.go":
		return true
	}
	if strings.HasSuffix(f, "_test.go") {
		return false
	}
	return false
}

// simple string hash for IDs
func hashID(s string) string {
	var h uint32
	for i := 0; i < len(s); i++ {
		h = h*31 + uint32(s[i])
	}
	return fmt.Sprintf("%08x", h)
}
