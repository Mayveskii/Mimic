package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mayveskii/Mimic/internal/mesh"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: migrate-gob-to-text <gob-dir> <output-dir>\n")
		os.Exit(1)
	}
	gobDir := os.Args[1]
	outDir := os.Args[2]

	fmt.Printf("Loading gob graphs from %s ...\n", gobDir)
	reg := mesh.NewRegistry()
	if err := reg.LoadAllGraphs(gobDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load graphs: %v\n", err)
		os.Exit(1)
	}

	// Load registry overlay if present
	registryPath := filepath.Join(gobDir, "..", "registry", "invariant_registry.json")
	if _, err := os.Stat(registryPath); err == nil {
		reg.LoadRegistry(registryPath)
	}

	fmt.Printf("Converting slots to text-native format ...\n")
	total := 0
	for _, graph := range reg.ListGraphs() {
		for _, node := range graph.Slots {
			slot := &mesh.TextSlot{
				ID:        node.ID,
				Domain:    graph.Domain,
				Invariant: node.Invariant,
				Embed:     node.EmbedInt8,
				Metadata: map[string]string{
					"source_repo": node.SourceRepo,
					"task":        node.Task,
					"usage_count": fmt.Sprintf("%d", node.UsageCount),
				},
			}
			// Convert ActionBytes to Actions if present
			if len(node.ActionBytes) > 0 {
				decoded, err := mesh.FormatActionBytes(node.ActionBytes)
				if err == nil {
					lines := splitLines(decoded)
					slot.Actions = lines
				}
			}
			// Context from source
			if node.WhyWorked != "" {
				slot.Context = node.WhyWorked
			}
			if node.Diff != "" {
				if slot.Context != "" {
					slot.Context += "\n\n"
				}
				slot.Context += "Diff:\n" + node.Diff
			}

			domainDir := filepath.Join(outDir, graph.Domain)
			if err := os.MkdirAll(domainDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", domainDir, err)
				continue
			}
			path := filepath.Join(domainDir, node.ID+".md")
			if err := os.WriteFile(path, slot.ToMarkdown(), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
				continue
			}
			total++
			if total%1000 == 0 {
				fmt.Printf("  ... %d slots converted\n", total)
			}
		}
	}
	fmt.Printf("Done. %d text-native slots written to %s\n", total, outDir)
}

func splitLines(s string) []string {
	var lines []string
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}
