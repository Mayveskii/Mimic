package mcp

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Mayveskii/Mimic/internal/embed"
	"github.com/Mayveskii/Mimic/internal/projectmap"
)

// ProjectMapHandler routes PROJECT_MAP_* tool calls.
type ProjectMapHandler struct {
	pm          *projectmap.ProjectMap
	embedClient *embed.Client
}

// NewProjectMapHandler opens or creates the project map for a workspace.
func NewProjectMapHandler(workingDir string, embedEndpoint string) *ProjectMapHandler {
	pm, err := projectmap.OpenOrCreate(workingDir)
	if err != nil {
		return &ProjectMapHandler{}
	}
	return &ProjectMapHandler{pm: pm, embedClient: embed.NewClient(embedEndpoint)}
}

func (h *ProjectMapHandler) Close() {
	if h.pm != nil {
		h.pm.Close()
	}
}

// HandleIndex triggers a full workspace reindex.
func (h *ProjectMapHandler) HandleIndex() map[string]interface{} {
	if h.pm == nil {
		return meshError("project map not initialized")
	}

	if err := h.pm.IndexWorkspace(); err != nil {
		return meshError("index failed: " + err.Error())
	}

	stats, _ := h.pm.Stats()
	b, _ := json.MarshalIndent(stats, "", "  ")
	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": string(b)}},
	}
}

// HandleStatus returns current project map stats.
func (h *ProjectMapHandler) HandleStatus() map[string]interface{} {
	if h.pm == nil {
		return meshError("project map not initialized")
	}

	stats, err := h.pm.Stats()
	if err != nil {
		return meshError("stats failed: " + err.Error())
	}
	b, _ := json.MarshalIndent(stats, "", "  ")
	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": string(b)}},
	}
}

// HandleQuerySymbol searches for symbols by prefix.
func (h *ProjectMapHandler) HandleQuerySymbol(args map[string]interface{}) map[string]interface{} {
	if h.pm == nil {
		return meshError("project map not initialized")
	}

	name, _ := args["name"].(string)
	if name == "" {
		return meshError("'name' is required")
	}

	syms, err := h.pm.QuerySymbol(name)
	if err != nil {
		return meshError("query failed: " + err.Error())
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d symbol(s) matching '%s':\n\n", len(syms), name)
	for i, s := range syms {
		if i >= 20 {
			fmt.Fprintf(&b, "... (%d more)\n", len(syms)-20)
			break
		}
		sig := ""
		if s.Signature != "" {
			sig = " | " + s.Signature
		}
		fmt.Fprintf(&b, "%s (%s) at %s:%d%s\n", s.Name, s.Type, s.File, s.Line, sig)
	}

	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": b.String()}},
	}
}

// HandleSearchText performs FTS5 search.
func (h *ProjectMapHandler) HandleSearchText(args map[string]interface{}) map[string]interface{} {
	if h.pm == nil {
		return meshError("project map not initialized")
	}

	query, _ := args["query"].(string)
	if query == "" {
		return meshError("'query' is required")
	}

	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	} else if l, ok := args["limit"].(int); ok {
		limit = l
	}

	results, err := h.pm.SearchText(query, limit)
	if err != nil {
		return meshError("search failed: " + err.Error())
	}

	var b strings.Builder
	fmt.Fprintf(&b, "FTS search '%s': %d result(s)\n\n", query, len(results))
	for i, r := range results {
		fmt.Fprintf(&b, "%d. %s (%s, %d bytes)\n   %s\n\n",
			i+1, r.Path, r.Lang, r.Size, r.Snippet)
	}

	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": b.String()}},
	}
}

// HandleSynthesize generates a mesh graph from the current workspace.
func (h *ProjectMapHandler) HandleSynthesize() map[string]interface{} {
	if h.pm == nil {
		return meshError("project map not initialized")
	}

	synth := projectmap.NewSynthesizer(h.pm, func(text string) [384]int8 {
		if h.embedClient == nil {
			return [384]int8{}
		}
		emb, err := h.embedClient.EmbedInt8(text)
		if err != nil {
			return [384]int8{}
		}
		return emb
	})

	sm, err := synth.Synthesize()
	if err != nil {
		return meshError("synthesize failed: " + err.Error())
	}

	result := map[string]interface{}{
		"domain":      sm.Name,
		"domain_id":   sm.Domain,
		"slots":       len(sm.Slots),
		"output_path": filepath.Join(h.pm.Root(), ".mimic", "workspace.graph.gob"),
		"status":      "success",
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	return map[string]interface{}{
		"content": []map[string]string{{"type": "text", "text": string(b)}},
	}
}

// NotifyFileChanged marks a file as dirty so the next operation reindexes it.
func (h *ProjectMapHandler) NotifyFileChanged(path string) {
	if h.pm == nil {
		return
	}
	_ = h.pm.IndexFile(path)
}
