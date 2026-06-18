package orchestrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mayveskii/Mimic/internal/config"
	"github.com/Mayveskii/Mimic/internal/graphify"
	"github.com/Mayveskii/Mimic/internal/mesh"
)

// MeshQuerier abstracts mesh slot retrieval for context assembly.
type MeshQuerier interface {
	QueryByFTS(query string, limit int) ([]*mesh.TextSlot, error)
	QueryByDomain(domain string) ([]*mesh.TextSlot, error)
}

// ContextAssembler builds the health pack consumed by the model.
type ContextAssembler struct {
	Mesh     MeshQuerier
	Graphify *graphify.Graph
	Config   *config.RepoConfig
}

// NewContextAssembler creates a context assembler for the given repo.
func NewContextAssembler(cfg *config.RepoConfig) *ContextAssembler {
	return &ContextAssembler{Config: cfg}
}

// Assemble builds a health pack string from intent, mesh, memory, graphify, and config.
func (a *ContextAssembler) Assemble(ctx context.Context, intent string) (string, error) {
	var parts []string

	// 1. Domain dialect activation
	domain := classifyDomain(intent)
	parts = append(parts, fmt.Sprintf("## Domain Dialect: %s\n", domain))

	// 2. Personas / primers / waivers
	if a.Config != nil {
		for _, p := range a.Config.Personas {
			content, err := loadFile(a.Config.Repo, p)
			if err == nil && content != "" {
				parts = append(parts, fmt.Sprintf("## Persona: %s\n%s\n", p, content))
			}
		}
		for _, p := range a.Config.Primers {
			content, err := loadFile(a.Config.Repo, p)
			if err == nil && content != "" {
				parts = append(parts, fmt.Sprintf("## Primer: %s\n%s\n", p, content))
			}
		}
	}

	// 3. Mesh health pack
	if a.Mesh != nil {
		slots, err := a.Mesh.QueryByFTS(intent, 5)
		if err == nil && len(slots) > 0 {
			parts = append(parts, "## Mesh Patterns\n")
			for _, slot := range slots {
				parts = append(parts, fmt.Sprintf("### Slot %s (%s)\n- Invariant: %s\n- Context: %s\n", slot.ID, slot.Domain, slot.Invariant, slot.Context))
				for _, link := range slot.Links {
					parts = append(parts, fmt.Sprintf("- Link: %s (%s, weight=%.2f)\n", link.TargetID, link.Relation, link.Weight))
				}
			}
		}
	}

	// 4. Graphify subgraph (if initialized)
	if a.Graphify != nil && len(a.Graphify.Nodes) > 0 {
		topNodes := a.Graphify.TopNodes(5)
		parts = append(parts, "## Repo Graph (Top Symbols)\n")
		for _, n := range topNodes {
			parts = append(parts, fmt.Sprintf("- %s %s (%s:%d) PageRank=%.4f\n", n.Type, n.Name, n.File, n.Line, n.PageRank))
		}
	}

	parts = append(parts, fmt.Sprintf("## Intent\n%s\n", intent))
	return strings.Join(parts, "\n"), nil
}

func classifyDomain(intent string) string {
	lower := strings.ToLower(intent)
	switch {
	case strings.Contains(lower, "git") || strings.Contains(lower, "commit") || strings.Contains(lower, "branch"):
		return "git"
	case strings.Contains(lower, "build") || strings.Contains(lower, "compile") || strings.Contains(lower, "test"):
		return "build"
	case strings.Contains(lower, "file") || strings.Contains(lower, "dir") || strings.Contains(lower, "path"):
		return "system"
	case strings.Contains(lower, "network") || strings.Contains(lower, "http") || strings.Contains(lower, "tcp"):
		return "network"
	case strings.Contains(lower, "process") || strings.Contains(lower, "spawn") || strings.Contains(lower, "kill"):
		return "process"
	case strings.Contains(lower, "mesh") || strings.Contains(lower, "pattern") || strings.Contains(lower, "invariant"):
		return "mesh"
	default:
		return "general"
	}
}

func loadFile(repo, path string) (string, error) {
	full := path
	if !filepath.IsAbs(path) {
		full = filepath.Join(repo, path)
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func repoPathFromConfig(cfg *config.RepoConfig) string {
	if cfg == nil {
		return "."
	}
	return cfg.Repo
}
